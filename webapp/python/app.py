import flask

app = flask.Flask(__name__)

@app.route("/initialize", methods=["POST"])
def post_initialize():
    return flask.jsonify({
        "language": "python" # 実装言語を返す
    })

if __name__ == "__main__":
    app.run(port=8000, debug=True, threaded=True)
