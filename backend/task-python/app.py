import json
import logging

from flask import Flask
from flask import make_response
from flask import request

from badcase_tagging import badcase_tagging_push

app = Flask(__name__)
logger = logging.getLogger('gunicorn.error')


@app.route('/badcase/push', methods=['POST', ])
def badcase_push():
    res_dict = {
        "code": 200,
        "error": "",
        "data": ""
    }

    if not request.content_type.startswith("application/json"):
        res_dict["code"] = 500
        res_dict["error"] = "not allowed content type!"
        return make_response(json.dumps(res_dict, ensure_ascii=False), res_dict["code"])

    data = request.get_data()
    json_re = json.loads(data)
    if json_re.__contains__("start_time"):
        start_time = json_re["start_time"]
    else:
        start_time = None

    if json_re.__contains__("end_time"):
        end_time = json_re["end_time"]
    else:
        end_time = None

    if json_re.__contains__("exclude_domain"):
        exclude_domain = json_re["exclude_domain"]
    else:
        exclude_domain = None

    if json_re.__contains__("exclude_intent"):
        exclude_intent = json_re["exclude_intent"]
    else:
        exclude_intent = None

    result, task_name = badcase_tagging_push(start_time=start_time, end_time=end_time,
                                             exclude_domain=exclude_domain, exclude_intent=exclude_intent)
    res_dict["code"] = result.status_code
    res_dict["data"] = result.json()

    return make_response(json.dumps(res_dict, ensure_ascii=False), res_dict["code"])


if __name__ != '__main__':
    gunicorn_logger = logging.getLogger('gunicorn.error')
    app.logger.handlers = gunicorn_logger.handlers
    app.logger.setLevel(gunicorn_logger.level)

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=8091)
