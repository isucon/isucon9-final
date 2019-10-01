import os
import sys
import datetime
import dateutil.parser
import logging

import flask
import requests
import MySQLdb.cursors

JST = datetime.timezone(datetime.timedelta(hours=+9), 'JST')

app = flask.Flask(__name__)

AvailableDays = 10
SessionName   = "session_isutrain"

TrainClassMap = {"express": "最速", "semi_express": "中間", "local": "遅いやつ"}


class HttpException(Exception):
    status_code = 500

    def __init__(self, status_code, message):
        Exception.__init__(self)
        self.message = message
        self.status_code = status_code

    def get_response(self):
        response = flask.jsonify({'is_error': True, 'message': self.message})
        response.status_code = self.status_code
        return response



def dbh():
    if not hasattr(flask.g, 'db'):
        flask.g.db = MySQLdb.connect(
            host=os.getenv('MYSQL_HOSTNAME', 'localhost'),
            port=int(os.getenv('MYSQL_PORT', 3306)),
            user=os.getenv('MYSQL_USER', 'isutrain'),
            password=os.getenv('MYSQL_PASSWORD', 'isutrain'),
            db=os.getenv('MYSQL_DATABASE', 'isutrain'),
            charset='utf8mb4',
            cursorclass=MySQLdb.cursors.DictCursor,
            autocommit=True,
        )
    return flask.g.db

def filter_dict_keys(d, allowed_keys):
    ret = {}
    for k, v in d.items():
        if k in allowed_keys:
            ret[k] = v
    return ret

@app.errorhandler(HttpException)
def handle_http_exception(error):
    return error.get_response()

def check_available_date(date):
    d = datetime.datetime(2020, 1, 1) + datetime.timedelta(days=AvailableDays)
    if d.date() <= date.date():
        return False
    return True


def get_usable_train_class_list(from_station, to_station):

    usable = TrainClassMap.values()

    for station in (from_station, to_station):
        if not station["is_stop_express"]:
            usable.remove(TrainClassMap["express"])

        if not station["is_stop_semi_express"]:
            usable.remove(TrainClassMap["semi_express"])

        if not station["is_stop_local"]:
            usable.remove(TrainClassMap["local"])

    return list(usable)


def get_available_seats_from_train(c, train, from_station, to_station, seat_class, is_smoking_seat):

    available_set_map = {}

    try:
        sql = "SELECT * FROM seat_master WHERE train_class=%s AND seat_class=%s AND is_smoking_seat=%s"

        c.execute(sql, (train["train_class"], seat_class, is_smoking_seat))
        seat_list = c.fetchall()

        for seat in seat_list:
            available_set_map["{}_{}_{}".format(seat["car_number"], seat["seat_row"], seat["seat_column"])] = seat

        sql = """SELECT sr.reservation_id, sr.car_number, sr.seat_row, sr.seat_column
        FROM seat_reservations sr, reservations r, seat_master s, station_master std, station_master sta
        WHERE
            r.reservation_id=sr.reservation_id AND
            s.train_class=r.train_class AND
            s.car_number=sr.car_number AND
            s.seat_column=sr.seat_column AND
            s.seat_row=sr.seat_row AND
            std.name=r.departure AND
            sta.name=r.arrival
        """

        if train["is_nobori"]:
            sql += " AND ((sta.id < %s AND %s <= std.id) OR (sta.id < %s AND %s <= std.id) OR (%s < sta.id AND std.id < %s))"
        else:
            sql += " AND ((std.id <= %s AND %s < sta.id) OR (std.id <= %s AND %s < sta.id) OR (sta.id < %s AND %s < std.id))"

        c.execute(sql, (from_station["id"], from_station["id"], to_station["id"], to_station["id"], from_station["id"], to_station["id"]))
        seat_reservation_list = c.fetchall()

        for seat_reservation in seat_reservation_list:
            key = "{}_{}_{}".format(seat_reservation["car_number"], seat_reservation["seat_row"], seat_reservation["seat_column"])
            if key in available_set_map:
                del(available_set_map[key])

    except MySQLdb.Error as err:
        app.logger.exception(err)
        raise HttpException(requests.codes['internal_server_error'], "db error")

    return  available_set_map.values()

def get_distance_fare(c, distance):

    sql = "SELECT distance,fare FROM distance_fare_master ORDER BY distance"
    c.execute(sql)

    distance_fare_list = c.fetchall()

    lastDistance = 0.0
    lastFare = 0
    for distanceFare in distance_fare_list:
        app.logger.warn("{} {} {}".format(distance, distanceFare["distance"], distanceFare["fare"]))
        if  lastDistance < distance and distance < distanceFare["distance"]:
            break
        lastDistance = distanceFare["distance"]
        lastFare = distanceFare["fare"]

    return lastFare

def calc_fare(c, date, from_station, to_station, train_class, seat_class):

    distance = abs(to_station["distance"] - from_station["distance"])
    distFare = get_distance_fare(c, distance)

    app.logger.warn("distFare {}".format(distFare))

    sql = "SELECT * FROM fare_master WHERE train_class=%s AND seat_class=%s ORDER BY start_date"
    c.execute(sql, (train_class, seat_class))
    fareList = c.fetchall()

    if len(fareList) == 0:
        raise HttpException(requests.codes['internal_server_error'], "fare_master does not exists")

    selectedFare = fareList[0]

    for fare in fareList:
        if fare["start_date"].date() <= date:
            app.logger.warn("%s %s", fare["start_date"].date(), fare["fare_multiplier"])
            selectedFare = fare

    app.logger.warn("%%%%%%%%%%%%%%%%%%%")
    return int(distFare * selectedFare["fare_multiplier"])

@app.route("/api/stations", methods=["GET"])
def get_stations():

    station_list = []

    try:
        conn = dbh()
        with conn.cursor() as c:
            sql = "SELECT * FROM `station_master` ORDER BY id"
            c.execute(sql)

            while True:
                station = c.fetchone()

                if station is None:
                    break

                station = filter_dict_keys(station, ["id", "name", "is_stop_express", "is_stop_semi_express", "is_stop_local"])
                station_list.append(station)

    except MySQLdb.Error as err:
        app.logger.exception(err)
        raise HttpException(requests.codes['internal_server_error'], "db error")

    return flask.jsonify(station_list)


@app.route("/api/train/search", methods=["GET"])
def get_train_search():

    use_at = dateutil.parser.parse(flask.request.args.get('use_at')).astimezone(JST)

    train_class = flask.request.args.get('train_class')
    from_name = flask.request.args.get('from')
    to_name = flask.request.args.get('to')

    adult = int(flask.request.args.get('adult'))
    child = int(flask.request.args.get('child'))

    if not check_available_date(use_at):
        raise HttpException(requests.codes['not_found'], "予約可能期間外です")

    trainSearchResponseList = []

    try:
        conn = dbh()
        with conn.cursor() as c:
            sql = "SELECT * FROM station_master WHERE name=%s"
            c.execute(sql, (from_name, ))
            from_station = c.fetchone()
            if not from_station:
                raise HttpException(requests.codes['bad_request'], "fromStation: no rows")


            c.execute(sql, (to_name, ))
            to_station = c.fetchone()
            if not to_station:
                raise HttpException(requests.codes['bad_request'], "toStation: no rows")


            is_nobori = False
            if from_station["distance"] > to_station["distance"]:
                is_nobori = True

            usable_train_class_list = get_usable_train_class_list(from_station, to_station)
            app.logger.warn("{}".format(usable_train_class_list))

            sql = "SELECT * FROM station_master ORDER BY distance"
            if is_nobori:
                # 上りだったら駅リストを逆にする
                query += " DESC"

            c.execute(sql)
            station_list = c.fetchall()

            if not train_class:
                sql = "SELECT * FROM train_master WHERE date=%s AND is_nobori=%s"
                c.execute(sql, (str(use_at.date()), is_nobori))
            else:
                sql = "SELECT * FROM train_master WHERE date=%s AND is_nobori=%s AND train_class=%s"
                c.execute(sql, (str(use_at.date()), is_nobori, train_class))

            train_search_response_list = []

            train_list = c.fetchall()

            for train in train_list:

                if train["train_class"] not in usable_train_class_list:
                    continue

                isSeekedToFirstStation = False
                isContainsOriginStation = False
                isContainsDestStation = False
                i = 0

                for station in station_list:

                    if not isSeekedToFirstStation:
                        # 駅リストを列車の発駅まで読み飛ばして頭出しをする
                        # 列車の発駅以前は止まらないので無視して良い
                        if station["name"] == train["start_station"]:
                            isSeekedToFirstStation = True
                        else:
                            continue

                    if station["id"] == from_station["id"]:
                        # 発駅を経路中に持つ編成の場合フラグを立てる
                        isContainsOriginStation = True


                    if station["id"] == to_station["id"]:
                        if isContainsOriginStation:
                            # 発駅と着駅を経路中に持つ編成の場合
                            isContainsDestStation = True
                            break
                        else:
                            # 出発駅より先に終点が見つかったとき
                            app.logger.warn("なんかおかしい")
                            break

                    if station["name"] == train["last_station"]:
                        # 駅が見つからないまま当該編成の終点に着いてしまったとき
                        break
                    i+=1

                if isContainsOriginStation and isContainsDestStation:
                    # 列車情報

                    sql = "SELECT departure FROM train_timetable_master WHERE date=%s AND train_class=%s AND train_name=%s AND station=%s"
                    c.execute(sql, (str(use_at.date()), train["train_class"], train["train_name"], from_station["name"]))
                    departure = c.fetchone()
                    departure = datetime.datetime(use_at.year, use_at.month, use_at.day, 0, 0, 0).replace(tzinfo=JST) + departure["departure"]

                    sql = "SELECT arrival FROM train_timetable_master WHERE date=%s AND train_class=%s AND train_name=%s AND station=%s"
                    c.execute(sql, (str(use_at.date()), train["train_class"], train["train_name"], to_station["name"]))
                    arrival = c.fetchone()
                    arrival = datetime.datetime(use_at.year, use_at.month, use_at.day, 0, 0, 0).replace(tzinfo=JST) + arrival["arrival"]


                    if use_at > departure:
                        # 乗りたい時刻より出発時刻が前なので除外
                        continue

                    premium_avail_seats = get_available_seats_from_train(c, train, from_station, to_station, "premium", False)
                    premium_smoke_avail_seats = get_available_seats_from_train(c, train, from_station, to_station, "premium", True)
                    reserved_avail_seats = get_available_seats_from_train(c, train, from_station, to_station, "reserved", False)
                    reserved_smoke_avail_seats = get_available_seats_from_train(c, train, from_station, to_station, "reserved", True)

                    premium_avail = "○"
                    if len(premium_avail_seats) == 0:
                        premium_avail = "×"
                    elif len(premium_avail_seats) < 10:
                        premium_avail = "△"

                    premium_smoke_avail = "○"
                    if len(premium_smoke_avail_seats) == 0:
                        premium_smoke_avail = "×"
                    elif len(premium_smoke_avail_seats) < 10:
                        premium_smoke_avail = "△"

                    reserved_avail = "○"
                    if len(reserved_avail_seats) == 0:
                        reserved_avail = "×"
                    elif len(reserved_avail_seats) < 10:
                        reserved_avail = "△"

                    reserved_smoke_avail = "○"
                    if len(reserved_smoke_avail_seats) == 0:
                        reserved_smoke_avail = "×"
                    elif len(reserved_smoke_avail_seats) < 10:
                        reserved_smoke_avail = "△"

                    # 空席情報
                    seatAvailability = {
                        "premium": premium_avail,
                        "premium_smoke": premium_smoke_avail,
                        "reserved": reserved_avail,
                        "reserved_smoke": reserved_smoke_avail,
                        "non_reserved": "○",
                    }

                    # 料金計算
                    premiumFare = calc_fare(c, use_at.date(), from_station, to_station, train["train_class"], "premium")
                    premiumFare = premiumFare*adult + premiumFare/2*child

                    reservedFare = calc_fare(c, use_at.date(), from_station, to_station, train["train_class"], "reserved")
                    reservedFare = reservedFare*adult + reservedFare/2*child

                    nonReservedFare = calc_fare(c, use_at.date(), from_station, to_station, train["train_class"], "non-reserved")
                    nonReservedFare = nonReservedFare*adult + nonReservedFare/2*child


                    fareInformation = {
                        "premium":        premiumFare,
                        "premium_smoke":  premiumFare,
                        "reserved":       reservedFare,
                        "reserved_smoke": reservedFare,
                        "non_reserved":   nonReservedFare,
                    }

                    trainSearchResponseList.append({
                        "train_class": train["train_class"],
                        "train_name": train["train_name"],
                        "start": train["start_station"],
                        "last": train["last_station"],
                        "departure": from_station["name"],
                        "arrival": to_station["name"],
                        "departure_time": str(departure.time()),
                        "arrival_time": str(arrival.time()),
                        "seat_availability": seatAvailability,
                        "seat_fare": fareInformation,
                    })

                    if len(trainSearchResponseList) >= 10:
                        break


    except MySQLdb.Error as err:
        app.logger.exception(err)
        raise HttpException(requests.codes['internal_server_error'], "db error")


    return flask.jsonify(trainSearchResponseList)


@app.route("/api/train/seats", methods=["GET"])
def get_train_seats():
    pass


@app.route("/api/train/reserve", methods=["POST"])
def post_reserve():
    pass

@app.route("/api/train/reservation/commit", methods=["POST"])
def post_commit():
    pass


@app.route("/api/auth", methods=["GET"])
def get_auth():
    pass

@app.route("/api/auth/signup", methods=["POST"])
def post_signup():
    pass


@app.route("/api/auth/login", methods=["POST"])
def post_login():
    pass

@app.route("/api/auth/logout", methods=["POST"])
def post_logout():
    pass

@app.route("/api/user/reservations", methods=["GET"])
def get_user_reservations():
    pass

@app.route("/api/user/reservations/:item_id", methods=["GET"])
def get_user_reservation_detail():
    pass


@app.route("/api/user/reservations/:item_id/cancel", methods=["POST"])
def post_user_reservation_cancel():
    pass


@app.route("/api/settings", methods=["GET"])
def get_settings():
    return flask.jsonify({
        "payment_api": os.getenv('PAYMENT_API', 'http://localhost:5000'),
    })


@app.route("/initialize", methods=["POST"])
def post_initialize():

    conn = dbh()
    with conn.cursor() as c:
        c.execute("TRUNCATE seat_reservations")
        c.execute("TRUNCATE reservations")
        c.execute("TRUNCATE users")

    return flask.jsonify({
        "language": "python", # 実装言語を返す
        "available_days": AvailableDays,
    })

if __name__ == "__main__":
    app.logger.setLevel(logging.DEBUG)
    app.run(port=8000, debug=True, threaded=True)
