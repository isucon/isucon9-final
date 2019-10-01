import flask
import requests
import os
import MySQLdb.cursors

app = flask.Flask(__name__)


AvailableDays = 10
SessionName   = "session_isutrain"


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


def http_json_error(code, msg):
    raise HttpException(code, msg)

@app.errorhandler(HttpException)
def handle_http_exception(error):
    return error.get_response()


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
        http_json_error(requests.codes['internal_server_error'], "db error")

    return flask.jsonify(station_list)


@app.route("/api/train/search", methods=["GET"])
def get_train_serch():
    pass


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
    app.run(port=8000, debug=True, threaded=True)
