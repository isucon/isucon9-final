import os
import sys
import datetime
import dateutil.parser
import logging

import flask
import pbkdf2
import requests
import MySQLdb.cursors

JST = datetime.timezone(datetime.timedelta(hours=+9), 'JST')

app = flask.Flask(__name__)
app.config['SECRET_KEY'] = 'isutrain'

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


def get_user():
    user_id = flask.session.get("user_id")
    if user_id is None:
        raise HttpException(requests.codes['unauthorized'], "no session")
    try:
        conn = dbh()
        with conn.cursor() as c:
            sql = "SELECT * FROM `users` WHERE `id` = %s"
            c.execute(sql, [user_id])
            user = c.fetchone()
            if user is None:
                raise HttpException(requests.codes['unauthorized'], "user not found")
    except MySQLdb.Error as err:
        app.logger.exception(err)
        http_json_error(requests.codes['internal_server_error'], "db error")
    return user


def filter_dict_keys(d, allowed_keys):
    ret = {}
    for k, v in d.items():
        if k in allowed_keys:
            ret[k] = v
    return ret

@app.errorhandler(HttpException)
def handle_http_exception(error):
    return error.get_response()

def message_response(message):
    return flask.jsonify({'is_error': False, 'message': message})

def check_available_date(date):
    d = datetime.datetime(2020, 1, 1) + datetime.timedelta(days=AvailableDays)
    if d.date() <= date:
        return False
    return True


def get_usable_train_class_list(from_station, to_station):

    usable = list(TrainClassMap.values())

    for station in (from_station, to_station):
        if not station["is_stop_express"] and TrainClassMap["express"] in usable:
            usable.remove(TrainClassMap["express"])

        if not station["is_stop_semi_express"] and TrainClassMap["semi_express"] in usable:
            usable.remove(TrainClassMap["semi_express"])

        if not station["is_stop_local"] and TrainClassMap["local"] in usable:
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


def make_reservation_response(c, reservation):
    sql = "SELECT departure FROM train_timetable_master WHERE date=%s AND train_class=%s AND train_name=%s AND station=%s"
    c.execute(sql, (
        reservation["date"],
        reservation["train_class"],
        reservation["train_name"],
        reservation["departure"]
    ))
    departure = c.fetchone()

    sql = "SELECT arrival FROM train_timetable_master WHERE date=%s AND train_class=%s AND train_name=%s AND station=%s"
    c.execute(sql, (
        reservation["date"],
        reservation["train_class"],
        reservation["train_name"],
        reservation["arrival"]
    ))
    arrival = c.fetchone()

    ret = filter_dict_keys(reservation,("reservation_id", "date", "amount", "adult", "child", "departure", "arrival", "train_class", "train_name"))
    reservation["departure_time"] = str(departure["departure"])
    reservation["arrival_time"] = str(arrival["arrival"])

    sql = "SELECT * FROM seat_reservations WHERE reservation_id=%s"
    c.execute(sql, (reservation["reservation_id"],))
    seat_reservation_list = c.fetchall()

    # 1つの予約内で車両番号は全席同じ
    reservation["car_number"] = seat_reservation_list[0]["car_number"]

    if reservation["car_number"] == 0:
        reservation["seat_class"] = "non-reserved"
    else:
        sql = "SELECT * FROM seat_master WHERE train_class=%s AND car_number=%s AND seat_column=%s AND seat_row=%s"
        c.execute(sql, (reservation["train_class"], reservation["car_number"], seat_reservation_list[0]["seat_column"], seat_reservation_list[0]["seat_row"]))
        seat = c.fetchone()
        reservation["seat_class"] = seat["seat_class"]


    reservation["seats"] = []
    for seat in seat_reservation_list:
        reservation["seats"].append({
            "seat_row": seat["seat_row"],
            "seat_column": seat["seat_column"],
        })

    return reservation


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
                station["is_stop_express"] = True if station["is_stop_express"] else False
                station["is_stop_semi_express"] = True if station["is_stop_semi_express"] else False
                station["is_stop_local"] = True if station["is_stop_local"] else False
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

    adult = int(flask.request.args.get('adult', '0'))
    child = int(flask.request.args.get('child', '0'))

    if not check_available_date(use_at.date()):
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
                sql += " DESC"

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
                    premiumFare = int(premiumFare*adult) + int(premiumFare/2*child)

                    reservedFare = calc_fare(c, use_at.date(), from_station, to_station, train["train_class"], "reserved")
                    reservedFare = int(reservedFare*adult) + int(reservedFare/2*child)

                    nonReservedFare = calc_fare(c, use_at.date(), from_station, to_station, train["train_class"], "non-reserved")
                    nonReservedFare = int(nonReservedFare*adult) + int(nonReservedFare/2*child)


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
    date = dateutil.parser.parse(flask.request.args.get('date')).astimezone(JST).date()

    train_class = flask.request.args.get('train_class')
    train_name = flask.request.args.get('train_name')
    from_name = flask.request.args.get('from')
    to_name = flask.request.args.get('to')

    car_number = int(flask.request.args.get('car_number'))

    if not check_available_date(date):
        raise HttpException(requests.codes['not_found'], "予約可能期間外です")


    seat_information_list = []
    car_list = []

    try:
        conn = dbh()
        with conn.cursor() as c:
            sql = "SELECT * FROM train_master WHERE date=%s AND train_class=%s AND train_name=%s"
            c.execute(sql, (str(date), train_class, train_name))
            train = c.fetchone()
            if not train:
                raise HttpException(requests.codes['not_found'], "列車が存在しません")


            sql = "SELECT * FROM station_master WHERE name=%s"
            c.execute(sql, (from_name, ))
            from_station = c.fetchone()
            if not from_station:
                raise HttpException(requests.codes['bad_request'], "fromStation: no rows")


            c.execute(sql, (to_name, ))
            to_station = c.fetchone()
            if not to_station:
                raise HttpException(requests.codes['bad_request'], "toStation: no rows")

            usable_train_class_list = get_usable_train_class_list(from_station, to_station)

            if train["train_class"] not in usable_train_class_list:
                raise HttpException(requests.codes['bad_request'], "invalid train_class")

            sql = "SELECT * FROM seat_master WHERE train_class=%s AND car_number=%s ORDER BY seat_row, seat_column"
            c.execute(sql, (train_class, car_number))

            seat_list = c.fetchall()

            for seat in seat_list:
                seat = {
                    "row": seat["seat_row"],
                    "column": seat["seat_column"],
                    "class": seat["seat_class"],
                    "is_smoking_seat": True if seat["is_smoking_seat"] else False,
                    "is_occupied": False,
                }

                sql = """
                SELECT s.*
                FROM seat_reservations s, reservations r
                WHERE
                    r.date=%s AND r.train_class=%s AND r.train_name=%s AND car_number=%s AND seat_row=%s AND seat_column=%s
                """

                c.execute(
                    sql,
                    (
                        str(date),
                        train_class,
                        train_name,
                        car_number,
                        seat["row"],
                        seat["column"],
                    )
                )

                seat_roweservation_list = c.fetchall()
                for seat_roweservation in seat_roweservation_list:
                    sql = "SELECT * FROM reservations WHERE reservation_id=%s"
                    c.execute(sql, (seat_reservation["reservation_id"],))
                    reservation = c.fetchone()

                    sql  = "SELECT * FROM station_master WHERE name=%s"
                    c.execute(sql, (reservation["departure"],))
                    departure_station = c.fetchone()
                    c.execute(sql, (reservation["arrival"],))
                    arrival_station = c.fetchone()

                    if train["is_nobori"]:
                        if to_station["id"] < arrival_station["id"] and from_station["id"] <= arrival_station["id"]:
                            pass
                        elif to_station["id"] >= departure_station["id"] and from_station["id"] > departure_station["id"]:
                            pass
                        else:
                            seat["is_occupied"] = True
                    else:
                        if from_station["id"] < departure_station["id"] and to_station["id"] <= departure_station["id"]:
                            pass
                        elif from_station["id"] >= arrival_station["id"] and to_station["id"] > arrival_station["id"]:
                            pass
                        else:
                            seat["is_occupied"] = True

                seat_information_list.append(seat)


            # 各号車の情報
            i = 1
            while True:
                sql = "SELECT * FROM seat_master WHERE train_class=%s AND car_number=%s ORDER BY seat_row, seat_column LIMIT 1"
                c.execute(sql, (train_class, i))
                seat = c.fetchone()
                if not seat:
                    break

                car_list.append({
                    "car_number": i,
                    "seat_class": seat["seat_class"],
                })

                i+=1

    except MySQLdb.Error as err:
        app.logger.exception(err)
        raise HttpException(requests.codes['internal_server_error'], "db error")


    return flask.jsonify({
        "date": str(date).replace("-", "/"),
        "train_class": train_class,
        "train_name": train_name,
        "car_number": car_number,
        "seats": seat_information_list,
        "cars":car_list
    })

@app.route("/api/train/reserve", methods=["POST"])
def post_reserve():
    body = flask.request.json

    date = dateutil.parser.parse(body.get('date')).astimezone(JST).date()

    app.logger.warn("%s", body)

    train_class = body.get('train_class')
    train_name = body.get('train_name')
    departure_name = body.get('departure')
    arrival_name = body.get('arrival')
    car_number = int(body.get('car_number'))
    seat_class = body.get('seat_class')
    is_smoking_seat = body.get('is_smoking_seat', False)

    adult = int(body.get('adult'))
    child = int(body.get('child'))

    column = body.get('column')
    seats =  body.get('seats', [])

    if not check_available_date(date):
        raise HttpException(requests.codes['not_found'], "予約可能期間外です")


    seat_information_list = []


    try:
        conn = dbh()
        with conn.cursor() as c:

            sql = "SELECT * FROM train_master WHERE date=%s AND train_class=%s AND train_name=%s"
            c.execute(sql, (str(date), train_class, train_name))
            train = c.fetchone()
            if not train:
                raise HttpException(requests.codes['not_found'], "列車が存在しません")


            sql = "SELECT * FROM station_master WHERE name=%s"
            c.execute(sql, (train["start_station"], ))
            start_station = c.fetchone()
            if not start_station:
                raise HttpException(requests.codes['not_found'], "リクエストされた列車の始発駅データがみつかりません")

            c.execute(sql, (train["last_station"], ))
            last_station = c.fetchone()
            if not last_station:
                raise HttpException(requests.codes['not_found'], "リクエストされた列車の終着駅データがみつかりません")


            c.execute(sql, (departure_name, ))
            from_station = c.fetchone()
            if not from_station:
                raise HttpException(requests.codes['not_found'], "リクエストされた乗車駅データがみつかりません")

            c.execute(sql, (arrival_name, ))
            to_station = c.fetchone()
            if not to_station:
                raise HttpException(requests.codes['not_found'], "リクエストされた降車駅データがみつかりません")

            usable_train_class_list = get_usable_train_class_list(from_station, to_station)

            if train["train_class"] not in usable_train_class_list:
                raise HttpException(requests.codes['bad_request'], "invalid train_class")


            # 運行していない区間を予約していないかチェックする
            if train["is_nobori"]:
                if from_station["id"] > start_station["id"] or to_station["id"] > start_station["id"]:
                    raise HttpException(requests.codes['bad_request'], "リクエストされた区間に列車が運行していない区間が含まれています")
                if last_station["id"] >= from_station["id"] or last_station["id"] > to_station["id"]:
                    raise HttpException(requests.codes['bad_request'], "リクエストされた区間に列車が運行していない区間が含まれています")
            else:
                if from_station["id"] < start_station["id"] or to_station["id"] < start_station["id"]:
                    raise HttpException(requests.codes['bad_request'], "リクエストされた区間に列車が運行していない区間が含まれています")
                if last_station["id"] <= from_station["id"] or last_station["id"] < to_station["id"]:
                    raise HttpException(requests.codes['bad_request'], "リクエストされた区間に列車が運行していない区間が含まれています")

            # あいまい座席検索
            # seatsが空白の時に発動する
            if not seats and seat_class != "non-reserved": #non-reservedはそもそもあいまい検索もせずダミーのRow/Columnで予約を確定させる。

                for car_number in range(1,17):
                    sql = "SELECT * FROM seat_master WHERE train_class=%s AND car_number=%s AND seat_class=%s AND is_smoking_seat=%s ORDER BY seat_row, seat_column"
                    c.execute(sql, (train_class, car_number, seat_class, is_smoking_seat))
                    seat_list = c.fetchall()
                    seats = [] # 予約対象席を空っぽに


                    for seat in seat_list:
                        sql = "SELECT s.* FROM seat_reservations s, reservations r WHERE r.date=%s AND r.train_class=%s AND r.train_name=%s AND car_number=%s AND seat_row=%s AND seat_column=%s FOR UPDATE"
                        c.execute(sql, (str(date), train_class, train_name, seat["car_number"], seat["seat_row"], seat["seat_column"]))
                        seat_reservation_list = c.fetchall()

                        is_occupied = False

                        for seat_reservation in seat_reservation_list:
                            sql = "SELECT * FROM reservations WHERE reservation_id=%s FOR UPDATE"
                            c.execute(sql, (seat_reservation["reservation_id"],))
                            reservation = c.fetchone()
                            if not reservation:
                                raise HttpException(requests.codes['bad_request'], "reservationが見つかりません")

                            sql = "SELECT * FROM station_master WHERE name=%s"
                            c.execute(sql, (reservation["departure"],))
                            departure_station = c.fetchone()
                            c.execute(sql, (reservation["arrival"],))
                            arrival_station = c.fetchone()

                            if train["is_nobori"]:
                                if to_station["id"] < arrival_station["id"] and from_station["id"] <= arrival_station["id"]:
                                    pass
                                elif to_station["id"] >= departure_station["id"] and from_station["id"] > departure_station["id"]:
                                    pass
                                else:
                                    is_occupied = True
                            else:
                                if from_station["id"] < departure_station["id"] and to_station["id"] <= departure_station["id"]:
                                    pass
                                elif from_station["id"] >= arrival_station["id"] and to_station["id"] > arrival_station["id"]:
                                    pass
                                else:
                                    is_occupied = True

                        seat_information_list.append({
                            "row": seat["seat_row"],
                            "column": seat["seat_column"],
                            "class": seat["seat_class"],
                            "is_smoking_seat": seat["is_smoking_seat"],
                            "is_occupied": is_occupied,
                        })

                    # 曖昧予約席とその他の候補席を選出
                    seatnum = adult + child - 1 #予約する座席の合計数 全体の人数からあいまい指定席分を引いておく
                    reserved = False #あいまい指定席確保済フラグ
                    vargue = True #あいまい検索フラグ
                    vague_seat = None #あいまい指定席保存用

                    if not column: #A/B/C/D/Eを指定しなければ、空いている適当な指定席を取るあいまいモード
                        seatnum = adult + child
                        reserved = True
                        vargue = False

                    candidate_seat_list = []

                    i = 0
                    for seat in seat_information_list:
                        if seat["column"] == column and not seat["is_occupied"] and not reserved and vargue: # あいまい席があいてる
                            vargue_seat = {
                                "row": seat["row"],
                                "column": seat["column"],
                            }
                            reserved = True
                        elif not seat["is_occupied"] and i < seatnum: #単に席があいてる
                            candidate_seat_list.append({
                                "row": seat["row"],
                                "column": seat["column"],
                            })
                            i+=1

                    if vargue and reserved: # あいまい席が見つかり、予約できそうだった
                        seats.append(vague_seat)
                    if i>0: # 候補席があった
                        seats += candidate_seat_list

                    if len(seats) < (adult + child):
                        # リクエストに対して席数が足りてない
                        # 次の号車にうつしたい
                        app.logger.warn("-----------------")
                        app.logger.warn("現在検索中の車両: %d号車, リクエスト座席数: %d, 予約できそうな座席数: %d, 不足数: %d", car_number, adult+child, len(seats), adult+child-len(seats))
                        app.logger.warn("リクエストに対して座席数が不足しているため、次の車両を検索します。")
                        if car_number == 16:
                            app.logger.warn("この新幹線にまとめて予約できる席数がなかったから検索をやめるよ")
                            raise HttpException(requests.codes['not_found'], "あいまい座席予約ができませんでした。指定した席、もしくは1車両内に希望の席数をご用意できませんでした。")

                    else:
                        app.logger.warn("空き実績: %d号車 シート:%s 席数:%d", car_number, seats, len(seats))
                        seats = seats[:adult+child]
                        break

            else:
                if len(seats) != (adult + child):
                    raise HttpException(requests.codes['bad_request'], "座席数が正しくありません")


            # 座席情報のValidate
            for seat in seats:
                sql = "SELECT * FROM seat_master WHERE train_class=%s AND car_number=%s AND seat_column=%s AND seat_row=%s AND seat_class=%s"
                c.execute(sql, (train_class, car_number, seat["column"], seat["row"], seat_class))
                if not c.fetchone():
                    raise HttpException(requests.codes['not_found'], "リクエストされた座席情報は存在しません。号車・喫煙席・座席クラスなど組み合わせを見直してください")


            # 当該列車・列車名の予約一覧取得
            sql = "SELECT * FROM reservations WHERE date=%s AND train_class=%s AND train_name=%s FOR UPDATE"
            c.execute(sql, (str(date), train_class, train_name))
            reservations = c.fetchall()

            for reservation in reservations:
                if seat_class == "non-reserved":
                    continue


                # 予約情報の乗車区間の駅IDを求める
                sql = "SELECT * FROM station_master WHERE name=%s"
                c.execute(sql, (reservation["departure"],))
                reservedfromStation = c.fetchone()
                if not reservedfromStation:
                    raise HttpException(requests.codes['internal_server_error'], "予約情報に記載された列車の乗車駅データがみつかりません")

                c.execute(sql, (reservation["arrival"],))
                reservedtoStation = c.fetchone()
                if not reservedtoStation:
                    raise HttpException(requests.codes['internal_server_error'], "予約情報に記載された列車の降車駅データがみつかりません")

                # 予約の区間重複判定
                secdup = False
                if train["is_nobori"]:
                    # 上り
                    if to_station["id"] < reservedtoStation["id"] and from_station["id"] <= reservedtoStation["id"]:
                        pass
                    elif to_station["id"] >= reservedfromStation["id"] and from_station["id"] > reservedfromStation["id"]:
                        pass
                    else:
                        secdup = True
                else:
                    # 下り
                    if from_station["id"] < reservedfromStation["id"] and to_station["id"] <= reservedfromStation["id"]:
                        pass
                    elif from_station["id"] >= reservedtoStation["id"] and to_station["id"] > reservedtoStation["id"]:
                        pass
                    else:
                        secdup = True

                if secdup:
                    # 区間重複の場合は更に座席の重複をチェックする
                    sql = "SELECT * FROM seat_reservations WHERE reservation_id=%s FOR UPDATE"
                    c.execute(sql, (reservation["reservation_id"],))
                    seat_reservations = c.fetchall()
                    for v in seat_reservations:
                        for seat in seats:
                            if v["car_number"] == car_number and v["seat_row"] == seat["row"] and v["seat_column"] == seat["column"]:
                                app.logger.warn("Duplicated ", reservation)
                                raise HttpException(requests.codes['bad_request'], "リクエストに既に予約された席が含まれています")

            # 3段階の予約前チェック終わり

            # 自由席は強制的にSeats情報をダミーにする（自由席なのに席指定予約は不可）
            if seat_class == "non-reserved":
                car_number = 0
                seats = []
                for num in range(adult, child):
                    seats.appaned({
                        "raw": 0,
                        "column": "",
                    })

            fare = calc_fare(c, date, from_station, to_station, train_class, seat_class)

            sumFare = int(adult * fare) + int(child*fare/2)
            app.logger.warn("SUMFARE %d", sumFare)

            # userID取得。ログインしてないと怒られる。
            user = get_user()


            # 予約ID発行と予約情報登録
            sql = "INSERT INTO `reservations` (`user_id`, `date`, `train_class`, `train_name`, `departure`, `arrival`, `status`, `payment_id`, `adult`, `child`, `amount`) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)"
            c.execute(
                sql,
                (
                    user["id"],
                    str(date),
                    train_class,
                    train_name,
                    departure_name,
                    arrival_name,
                    "requesting",
                    "a",
                    adult,
                    child,
                    sumFare,
                )
            )
            reservation_id = c.lastrowid

            # 席の予約情報登録
            # reservationsレコード1に対してseat_reservationstが1以上登録される
            sql = "INSERT INTO `seat_reservations` (`reservation_id`, `car_number`, `seat_row`, `seat_column`) VALUES (%s, %s, %s, %s)"
            for seat in seats:
                c.execute(sql, (reservation_id, car_number, seat["row"], seat["column"]))


    except MySQLdb.Error as err:
        conn.rollback()
        app.logger.exception(err)
        raise HttpException(requests.codes['internal_server_error'], "db error")
    except:
        conn.rollback()
        raise

    return flask.jsonify({
        "reservation_id": reservation_id,
        "amount": sumFare,
        "is_ok": True,
    })

@app.route("/api/train/reservation/commit", methods=["POST"])
def post_commit():
    app.logger.warn("XXXXX1 %s", flask.request)
    app.logger.warn("XXXXX2 %s", flask.request.form)
    app.logger.warn("XXXXX3 %s", flask.request.json)
    body = flask.request.json
    app.logger.warn("XXXX %s", body)

    reservation_id = body.get('reservation_id')
    card_token = body.get('card_token')

    user = get_user()

    app.logger.warn("/api/train/reservation/commit")

    try:
        conn = dbh()
        with conn.cursor() as c:
            # 予約IDで検索
            sql = "SELECT * FROM reservations WHERE reservation_id=%s"
            c.execute(sql, (reservation_id,))
            reservation = c.fetchone()
            if not reservation:
                raise HttpException(requests.codes['not_found'], "予約情報がみつかりません")


            if reservation["user_id"] != user["id"]:
                raise HttpException(requests.codes['forbidden'], "他のユーザIDの支払いはできません")

            if reservation["status"] == "done":
                raise HttpException(requests.codes['forbidden'], "既に支払いが完了している予約IDです")

            payment_api = os.getenv('PAYMENT_API', 'http://payment:5000')

            res = requests.post(payment_api+"/payment", json={
                "payment_information":{
                    "card_token": card_token,
                    "reservation_id": reservation["reservation_id"],
                    "amount": reservation["amount"],
                }
            })

            if res.status_code != 200:
                raise HttpException(requests.codes['internal_server_error'], "決済に失敗しました。カードトークンや支払いIDが間違っている可能性があります")

            payment_res = res.json()

            # 予約情報の更新
            sql = "UPDATE reservations SET status=%s, payment_id=%s WHERE reservation_id=%s"
            c.execute(sql, ("done", payment_res["payment_id"], reservation["reservation_id"]))

    except MySQLdb.Error as err:
        conn.rollback()
        app.logger.exception(err)
        raise HttpException(requests.codes['internal_server_error'], "db error")
    except:
        conn.rollback()
        raise

    return flask.jsonify({"is_ok": True})


@app.route("/api/auth", methods=["GET"])
def get_auth():
    user = get_user()
    return flask.jsonify({
        "email": user["email"],
    })

@app.route("/api/auth/signup", methods=["POST"])
def post_signup():
    email = flask.request.json['email']
    password = flask.request.json['password']
    super_secure_password = pbkdf2.crypt(password, iterations=100)

    try:
        conn = dbh()
        with conn.cursor() as c:
            sql = "INSERT INTO `users` (`email`, `salt`, `super_secure_password`) VALUES (%s, %s, %s)"
            c.execute(sql, (email, "", super_secure_password))

    except MySQLdb.Error as err:
        app.logger.exception(err)
        raise HttpException(requests.codes['internal_server_error'], "db error")

    return message_response("registration complete")

@app.route("/api/auth/login", methods=["POST"])
def post_login():
    email = flask.request.json['email']
    password = flask.request.json['password']


    try:
        conn = dbh()
        with conn.cursor() as c:
            sql = "SELECT * FROM users WHERE email=%s"
            c.execute(sql, (email,))
            user = c.fetchone()
            if not user:
                raise HttpException(requests.codes['forbidden'], "authentication failed")

            if pbkdf2.crypt(password, user["super_secure_password"]) != user["super_secure_password"].decode("ascii"):
                raise HttpException(requests.codes['forbidden'], "authentication failed")

            flask.session['user_id'] = user["id"]

    except MySQLdb.Error as err:
        app.logger.exception(err)
        raise HttpException(requests.codes['internal_server_error'], "db error")

    return message_response("ok")

@app.route("/api/auth/logout", methods=["POST"])
def post_logout():
    if "user_id" in flask.session:
        del(flask.session['user_id'])
    return message_response("ok")

@app.route("/api/user/reservations", methods=["GET"])
def get_user_reservations():
    user = get_user()

    ret = []
    try:
        conn = dbh()
        with conn.cursor() as c:
            sql = "SELECT * FROM reservations WHERE user_id=%s"
            c.execute(sql, (user["id"],))
            reservations = c.fetchall()

            for reservation in reservations:
                ret.append(make_reservation_response(c, reservation))

    except MySQLdb.Error as err:
        app.logger.exception(err)
        raise HttpException(requests.codes['internal_server_error'], "db error")


    return flask.jsonify(ret)

@app.route("/api/user/reservations/<item_id>", methods=["GET"])
def get_user_reservation_detail(item_id):
    user = get_user()

    reservation = None
    try:
        conn = dbh()
        with conn.cursor() as c:
            sql = "SELECT * FROM reservations WHERE user_id=%s AND reservation_id=%s"
            c.execute(sql, (user["id"], item_id))
            reservation = c.fetchone()
            if not reservation:
                raise HttpException(requests.codes['not_found'], "Reservation not found")

            reservation = make_reservation_response(c, reservation)

    except MySQLdb.Error as err:
        app.logger.exception(err)
        raise HttpException(requests.codes['internal_server_error'], "db error")

    return flask.jsonify(reservation)



@app.route("/api/user/reservations/<item_id>/cancel", methods=["POST"])
def post_user_reservation_cancel(item_id):
    user = get_user()

    reservation = None
    try:
        conn = dbh()
        with conn.cursor() as c:
            sql = "SELECT * FROM reservations WHERE user_id=%s AND reservation_id=%s"
            c.execute(sql, (user["id"], item_id))
            reservation = c.fetchone()
            if not reservation:
                raise HttpException(requests.codes['not_found'], "Reservation not found")


            if reservation["status"] == "rejected":
                raise HttpException(requests.codes['internal_server_error'], "何らかの理由により予約はRejected状態です")

            if reservation["status"] == "done":

                payment_api = os.getenv('PAYMENT_API', 'http://payment:5000')

                res = requests.delete(payment_api+"/payment/" + reservation["payment_id"])

                if res.status_code != 200:
                    raise HttpException(requests.codes['internal_server_error'], "決済のキャンセルに失敗しました")

            sql = "DELETE FROM reservations WHERE reservation_id=%s"
            c.execute(sql, (reservation["reservation_id"],))

            sql = "DELETE FROM seat_reservations WHERE reservation_id=%s"
            c.execute(sql, (reservation["reservation_id"],))



    except MySQLdb.Error as err:
        conn.rollback()
        app.logger.exception(err)
        raise HttpException(requests.codes['internal_server_error'], "db error")
    except:
        conn.rollback()
        raise

    return message_response("cancel completed")



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
