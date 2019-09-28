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
        "language": "python" # 実装言語を返す
        "available_days": AvailableDays,
    })

if __name__ == "__main__":
    app.run(port=8000, debug=True, threaded=True)
