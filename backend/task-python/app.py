import json
import logging

from flask import Flask
from flask import make_response
from flask import request
from flask_apscheduler import APScheduler

from badcase_tagging import badcase_tagging_push, badcase_tagging_pull


class Config(object):
    JOBS = [
        {
            "id": "bad_case_pull",
            "func": "__main__:badcase_tagging_pull",
            "trigger": "cron",
            "day_of_week": 1,  # 每周二
            "hour": 9,  # 早上9点
            "minute": 30,  # 30分
        },
        {
            "id": "bad_case_push",
            "func": "__main__:badcase_tagging_push",
            "trigger": "cron",
            "day_of_week": 1,  # 每周二
            "hour": 9,  # 早上9点
            "minute": 30,  # 30分
        }
    ]
    SCHEDULER_API_ENABLED = True
    SCHEDULER_TIMEZONE = 'Asia/Shanghai'


app = Flask(__name__)
logger = logging.getLogger('gunicorn.error')

app.config.from_object(Config())
scheduler = APScheduler()
scheduler.init_app(app)
scheduler.start()


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


@app.route('/badcase/pull', methods=['GET', ])
def bad_case_pull():
    res_dict = {
        "code": 200,
        "error": "",
        "data": ""
    }
    ids, cases = badcase_tagging_pull()
    if not cases:
        res_dict["error"] = "no more cases available"
        res_dict["code"] = 500
        return make_response(json.dumps(res_dict, ensure_ascii=False), res_dict["code"])

    res_dict["data"] = f"add test cases successfully! {len(cases)}/{len(ids)}"
    return make_response(json.dumps(res_dict, ensure_ascii=False), res_dict["code"])


if __name__ != '__main__':
    gunicorn_logger = logging.getLogger('gunicorn.error')
    app.logger.handlers = gunicorn_logger.handlers
    app.logger.setLevel(gunicorn_logger.level)

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=8091)
