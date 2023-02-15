# -*- coding:utf-8 -*-
import json
import logging

from flask import Flask
from flask import make_response
from flask import request
from google.protobuf import json_format

# from flask_apscheduler import APScheduler
# from apscheduler.schedulers.background import BackgroundScheduler

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

# app.config.from_object(Config())
# scheduler = APScheduler(BackgroundScheduler())
# scheduler.init_app(app)
# scheduler.start()


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


@app.route('/interface', methods=['POST', ])
def interface_test():
    import traceback

    def template(proto, url, payload, stub, req_func, call_func, iterator):
        return f"""
    # -*- coding:utf-8 -*-
import {proto}_pb2 as pb2
import {proto}_pb2_grpc as pb2_grpc
import json
import grpc
from google.protobuf import json_format


def pb_to_json(pb):
    return json_format.MessageToJson(pb)


def json_to_pb(json_obj, message):
    return json_format.Parse(json.dumps(json_obj, indent=4), message)


def yield_message(message):
    yield message


class Interface:
    def __init__(self, url, stub):
        self.channel = grpc.insecure_channel(url)
        self.stub = stub(self.channel)

    def __del__(self):
        self.channel.close()

    @staticmethod
    def call(payload, message, func, iterator=False):
        request = yield_message(json_to_pb(payload, message)) if iterator else json_to_pb(payload, message)
        return list(func(request))


ins = Interface(url={url}, stub=pb2_grpc.{stub})
result = ins.call(message=pb2.{req_func}(), func=ins.stub.{call_func}, payload={payload}, iterator={iterator})
proResult["result"] = result
        """
    if not request.content_type.startswith("application/json"):
        return make_response(json.dumps({"error": "not allowed content type!"}, ensure_ascii=False), 500)

    data = request.get_data()
    json_re = json.loads(data)

    def pb_to_json(pb):
        return json_format.MessageToJson(pb)
    import os
    proto_file = json_re["proto"] if json_re.__contains__("proto") else "abc.proto"
    proto_file_pb2 = proto_file.split(".")[0] + "_pb2.py"
    proto_file_pb2_grpc = proto_file.split(".")[0] + "_pb2_grpc.py"

    if not os.path.isfile(proto_file_pb2) or not os.path.isfile(proto_file_pb2_grpc):
        if not os.path.isfile(proto_file):
            return make_response({"error": f"not find {proto_file}"}, 200)
        cmd = f'python -m grpc_tools.protoc -I. --python_out=. --grpc_python_out=. {proto_file}'
        result = os.system(cmd)
        if result != 0:
            return make_response({"error": f'error when execute command: {cmd}', "code": result}, 200)
    url = json_re["url"]
    proResult = {"result": []}
    sc = template(proto=proto_file.split(".")[0],
                  url=f"'{url}'",
                  payload=json_re["payload"],
                  stub=json_re["stub"],
                  req_func=json_re["req_func"],
                  call_func=json_re["call_func"],
                  iterator=json_re["iterator"])
    try:
        exec(sc, {"proResult": proResult})
    except Exception as e:
        return make_response({"error": traceback.format_exc()}, 200)
    results = []
    for r in proResult["result"]:
        results.append(json.loads(pb_to_json(r)))
    return make_response(results, 200)


if __name__ != '__main__':
    gunicorn_logger = logging.getLogger('gunicorn.error')
    app.logger.handlers = gunicorn_logger.handlers
    app.logger.setLevel(gunicorn_logger.level)

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=8091)
