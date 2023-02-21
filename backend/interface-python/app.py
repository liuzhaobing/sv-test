# -*- coding:utf-8 -*-
import json
import logging
import os

from flask import Flask
from flask import make_response
from flask import request
from google.protobuf import json_format

app = Flask(__name__)
logger = logging.getLogger('gunicorn.error')

UPLOAD_FOLDER = ''
ALLOWED_EXTENSIONS = {'proto'}

app.config['UPLOAD_FOLDER'] = UPLOAD_FOLDER


def allowed_file(filename):
    return '.' in filename and filename.rsplit('.', 1)[1] in ALLOWED_EXTENSIONS


@app.route('/files', methods=['POST', ])
def upload():
    """{"file": file.proto}"""
    if request.method == 'POST':
        # 检查文件是否已经上传
        if 'file' not in request.files:
            return make_response({"status": "failure", "error": "upload file failed!"}, 500)

        file = request.files['file']
        # 如果用户没有选择文件，则提交一个空表单
        if file.filename == '':
            return make_response({"status": "failure", "error": "not find file!"}, 500)

        # 如果文件是允许的类型，则保存文件
        if file and allowed_file(file.filename):
            file.save(file.filename)
            return make_response({"status": "success", "error": ""}, 200)
        return make_response({"status": "failure", "error": "not support file type!"}, 500)


@app.route('/files', methods=['GET', ])
def list_proto_files():
    files = os.listdir()
    return make_response({"status": "success", "data": [i for i in files if allowed_file(i)]}, 200)


def generate_pb(proto_file):
    cmd = f'python -m grpc_tools.protoc -I. -I./src --python_out=. --grpc_python_out=. {proto_file}'
    return os.system(cmd), cmd


@app.route('/generate', methods=['POST', ])
def generate_pb_interface():
    data = request.get_data()
    json_re = json.loads(data)
    if not allowed_file(json_re["filename"]):
        return make_response({"status": "failure", "error": f'type error! only support "*.proto"'}, 500)

    if not json_re.__contains__("filename"):
        return make_response({"status": "failure", "error": f'key error! only support key "filename"'}, 500)

    if not os.path.exists(json_re["filename"]):
        return make_response({"status": "failure", "error": f'not find file: {json_re["filename"]}'}, 500)

    result, cmd = generate_pb(json_re["filename"])
    if result != 0:
        return make_response({"status": "failure", "error": f'execute command error: {cmd}', "code": result}, 500)
    return make_response({"status": "success", "error": ""}, 200)


@app.route('/interface', methods=['POST', ])
def interface_test():
    import traceback

    def template(proto, url, payload, stub, req_func, call_func, request_stream, response_stream):
        return f"""
    # -*- coding:utf-8 -*-
import {proto}_pb2 as pb2
import {proto}_pb2_grpc as pb2_grpc
import json
import grpc
import time
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
    def call(payload, message, func, request_stream=False, response_stream=False):
        request = yield_message(json_to_pb(payload, message)) if request_stream else json_to_pb(payload, message)
        return list(func(request)) if response_stream else func(request)


ins = Interface(url={url}, stub=pb2_grpc.{stub})
start_time = time.time()
result = ins.call(message=pb2.{req_func}(), func=ins.stub.{call_func}, payload={payload}, request_stream={request_stream}, response_stream={response_stream})
exec_result["result"] = result
exec_result["cost"] = int((time.time() - start_time) * 1000)
        """
    if not request.content_type.startswith("application/json"):
        return make_response(json.dumps({"status": "failure", "error": "not allowed content type!"}, ensure_ascii=False), 500)

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
            return make_response({"status": "failure", "error": f"not find {proto_file}"}, 200)
        result, cmd = generate_pb(proto_file)
        if result != 0:
            return make_response({"status": "failure", "error": f'execute command error: {cmd}', "code": result}, 200)
    url = json_re["url"]
    res = {"result": []}
    func = json_re["func"]

    sc = template(proto=proto_file.split(".")[0],
                  url=f"'{url}'",
                  payload=json_re["payload"],
                  stub=func.split(".")[0],
                  req_func=func.split(".")[1].split("(")[1].split(")")[0],
                  call_func=func.split(".")[1].split("(")[0],
                  request_stream=json_re["stream_in"],
                  response_stream=json_re["stream_out"])
    try:
        exec(sc, {"exec_result": res})
    except:
        return make_response({"status": "failure", "error": traceback.format_exc()}, 200)
    if json_re["stream_out"] == "False":
        return make_response({"cost": res["cost"], "data": json.loads(pb_to_json(res["result"]))}, 200)
    return make_response({"cost": res["cost"], "data": [json.loads(pb_to_json(r)) for r in res["result"]]}, 200)


if __name__ != '__main__':
    gunicorn_logger = logging.getLogger('gunicorn.error')
    app.logger.handlers = gunicorn_logger.handlers
    app.logger.setLevel(gunicorn_logger.level)

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=27998)
