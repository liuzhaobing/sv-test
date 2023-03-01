# -*- coding:utf-8 -*-
import json
import logging
import os
from datetime import datetime

from flask import Flask
from flask import make_response
from flask import request
from google.protobuf import json_format
from flask_sqlalchemy import SQLAlchemy

app = Flask(__name__)
logger = logging.getLogger('gunicorn.error')

UPLOAD_FOLDER = ''
ALLOWED_EXTENSIONS = {'proto'}

app.config['UPLOAD_FOLDER'] = UPLOAD_FOLDER
app.config["SQLALCHEMY_DATABASE_URI"] = "mysql+pymysql://root@172.16.23.33:3306/nlpautotest"
app.config["SQLALCHEMY_TRACK_MODIFICATIONS"] = True
app.config["SQLALCHEMY_COMMIT_ON_TEARDOWN"] = True
db = SQLAlchemy(app)


class ProtoManagement(db.Model):
    __tablename__ = "proto_management"
    id = db.Column(db.Integer, autoincrement=True, primary_key=True)
    proto_name_zh = db.Column(db.String(255))
    proto_name_en = db.Column(db.String(255))
    proto_content = db.Column(db.Text)
    last_access_time = db.Column(db.DATETIME, default=datetime.now())

    def to_dict(self):
        return {
            "id": self.id,
            "proto_name_zh": self.proto_name_zh,
            "proto_name_en": self.proto_name_en,
            "proto_content": self.proto_content,
            "last_access_time": self.last_access_time.strftime("%Y-%m-%d %H:%M:%S")
        }


class ProtoManagementLog(db.Model):
    __tablename__ = "proto_management_log"
    id = db.Column(db.Integer, autoincrement=True, primary_key=True)
    proto_name_en = db.Column(db.String(255))
    last_access_address = db.Column(db.String(255))
    last_access_function = db.Column(db.String(255))
    last_access_request = db.Column(db.Text)
    last_access_response = db.Column(db.Text)
    last_access_time = db.Column(db.DATETIME, default=datetime.now())
    stream_in = db.Column(db.String(255))
    stream_out = db.Column(db.String(255))

    def to_dict(self):
        return {
            "id": self.id,
            "proto_name_en": self.proto_name_en,
            "last_access_address": self.last_access_address,
            "last_access_function": self.last_access_function,
            "last_access_request": self.last_access_request,
            "last_access_response": self.last_access_response,
            "last_access_time": self.last_access_time.strftime("%Y-%m-%d %H:%M:%S"),
            "stream_in": self.stream_in,
            "stream_out": self.stream_out
        }


@app.route('/proto', methods=['GET', 'POST'])
def proto_management():
    if request.method == "GET":
        result = ProtoManagement.query.order_by(ProtoManagement.last_access_time.desc())
        return make_response({"status": "success", "data": [item.to_dict() for item in result]}, 200)
    if request.method == "POST":
        data = request.get_data()
        json_re = json.loads(data)
        check_proto_use_history = ProtoManagement.query.filter_by(proto_name_en=json_re["proto_name_en"]).count()
        write_local_file(json_re["proto_name_en"], json_re["proto_content"])
        gen_res, cmd = generate_pb(json_re["proto_name_en"])
        if gen_res != 0:
            # 判断pb文件生成是否成功
            return make_response({"status": "failure", "error": f'execute command error: {cmd}', "code": gen_res}, 200)

        if not check_proto_use_history:
            # 判断之前是否有存储过此proto文件
            new_proto = ProtoManagement(proto_name_zh=json_re["proto_name_zh"],
                                        proto_name_en=json_re["proto_name_en"],
                                        proto_content=json_re["proto_content"])
            result = db.session.add(new_proto)
            return make_response({"status": "success", "data": result}, 200)

        result = ProtoManagement.query.filter_by(proto_name_en=json_re["proto_name_en"]).update(json_re)
        return make_response({"status": "success", "data": result}, 200)


@app.route('/proto/<int:nid>', methods=['GET', 'PUT', 'DELETE'])
def proto_management_crud(nid):
    if request.method == "GET":
        result = ProtoManagement.query.filter_by(id=nid).first()
        return make_response({"status": "success", "data": result.to_dict()}, 200)

    if request.method == "DELETE":
        result = ProtoManagement.query.filter_by(id=nid).delete()
        return make_response({"status": "success", "data": result}, 200)

    if request.method == "PUT":
        data = request.get_data()
        json_re = json.loads(data)
        write_local_file(json_re["proto_name_en"], json_re["proto_content"])
        gen_res, cmd = generate_pb(json_re["proto_name_en"])
        if gen_res != 0:
            # 判断pb文件生成是否成功
            return make_response({"status": "failure", "error": f'execute command error: {cmd}', "code": gen_res}, 200)
        result = ProtoManagement.query.filter_by(id=nid).update(json_re)
        return make_response({"status": "success", "data": result}, 200)


@app.route('/proto/logs', methods=['GET', 'POST'])
def proto_management_log():
    if request.method == "GET":
        result = ProtoManagementLog.query.order_by(ProtoManagementLog.last_access_time.desc())
        return make_response({"status": "success", "data": [item.to_dict() for item in result]}, 200)
    if request.method == "POST":
        data = request.get_data()
        json_re = json.loads(data)
        new_proto = ProtoManagementLog(proto_name_en=json_re["proto_name_en"],
                                       last_access_address=json_re["last_access_address"],
                                       last_access_function=json_re["last_access_function"],
                                       last_access_request=json_re["last_access_request"],
                                       last_access_response=json_re["last_access_response"],
                                       stream_in=json_re["stream_in"],
                                       stream_out=json_re["stream_out"])
        result = db.session.add(new_proto)
        return make_response({"status": "success", "data": result}, 200)


@app.route('/proto/logs/<proto_name_en>', methods=['GET', 'PUT', 'DELETE'])
def proto_management_log_crud(proto_name_en):
    if request.method == "GET":
        result = ProtoManagementLog.query.filter_by(proto_name_en=proto_name_en).order_by(ProtoManagementLog.last_access_time.desc())
        return make_response({"status": "success", "data": [item.to_dict() for item in result]}, 200)

    if request.method == "DELETE":
        result = ProtoManagementLog.query.filter_by(proto_name_en=proto_name_en).delete()
        return make_response({"status": "success", "data": result}, 200)

    if request.method == "PUT":
        data = request.get_data()
        json_re = json.loads(data)
        result = ProtoManagementLog.query.filter_by(proto_name_en=proto_name_en).update(json_re)
        return make_response({"status": "success", "data": result}, 200)


def allowed_file(filename):
    return '.' in filename and filename.rsplit('.', 1)[1] in ALLOWED_EXTENSIONS


def write_local_file(filename: str, file_content: str):
    with open(filename, "w") as f:
        f.write(file_content)


def check_file_exists(filename: str) -> bool:
    return os.path.exists(filename)


def generate_pb(proto_file):
    cmd = f'python -m grpc_tools.protoc -I. -I./src --python_out=. --grpc_python_out=. {proto_file}'
    return os.system(cmd), cmd


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
        return make_response(
            json.dumps({"status": "failure", "error": "not allowed content type!"}, ensure_ascii=False), 500)

    data = request.get_data()
    json_re = json.loads(data)

    def pb_to_json(pb):
        return json_format.MessageToJson(pb)

    proto_file = json_re["proto"] if json_re.__contains__("proto") else "abc.proto"
    proto_file_pb2 = proto_file.split(".")[0] + "_pb2.py"
    proto_file_pb2_grpc = proto_file.split(".")[0] + "_pb2_grpc.py"

    if not os.path.isfile(proto_file_pb2) or not os.path.isfile(proto_file_pb2_grpc):
        if not os.path.isfile(proto_file):
            check_proto_use_history = ProtoManagement.query.filter_by(proto_name_en=json_re["proto"]).count()
            if not check_proto_use_history:
                return make_response({"status": "failure", "error": f"not find {proto_file}"}, 200)
            result = ProtoManagement.query.filter_by(proto_name_en=json_re["proto"]).first()
            write_local_file(json_re["proto"], result.proto_content)
        result, cmd = generate_pb(proto_file)
        if result != 0:
            return make_response({"status": "failure", "error": f'execute command error: {cmd}', "code": result}, 200)
    url = json_re["url"]
    res = {"result": []}
    func = json_re["func"]

    sc = template(proto=proto_file.split(".")[0],
                  url=f"'{url}'",
                  payload=json_re["payload"],
                  stub=func.split(".")[0] + "Stub",
                  req_func=func.split(".")[1].split("(")[1].split(")")[0],
                  call_func=func.split(".")[1].split("(")[0],
                  request_stream=json_re["stream_in"],
                  response_stream=json_re["stream_out"])
    try:
        exec(sc, {"exec_result": res})
        response_body = json.loads(pb_to_json(res["result"])) if json_re["stream_out"] == "False" \
            else [json.loads(pb_to_json(r)) for r in res["result"]]
        response_json = {"cost": res["cost"], "data": response_body}
        database_log = ProtoManagementLog()
        database_log.proto_name_en = json_re["proto"]
        database_log.last_access_address = json_re["url"]
        database_log.last_access_function = func
        database_log.last_access_request = json.dumps(json_re["payload"], ensure_ascii=False)
        database_log.last_access_response = json.dumps(response_json, ensure_ascii=False)
        database_log.stream_in = json_re["stream_in"]
        database_log.stream_out = json_re["stream_out"]
        db.session.add(database_log)
        return make_response(response_json, 200)
    except:
        response_json = {"status": "failure", "error": traceback.format_exc()}
        return make_response(response_json, 200)


if __name__ != '__main__':
    gunicorn_logger = logging.getLogger('gunicorn.error')
    app.logger.handlers = gunicorn_logger.handlers
    app.logger.setLevel(gunicorn_logger.level)

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=27998)
