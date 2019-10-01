import flask
import os
import MySQLdb.cursors

app = flask.Flask(__name__)


AvailableDays = 10
SessionName   = "session_isutrain"


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


@app.route("/api/stations", methods=["GET"])
def get_stations():
    pass


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
