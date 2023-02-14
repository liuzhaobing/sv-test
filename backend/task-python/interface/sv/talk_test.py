# -*- coding:utf-8 -*-
import uuid

import json
import grpc
import talk_pb2
import talk_pb2_grpc

from google.protobuf import json_format


def pb_to_json(pb):
    """将google.protobuf.struct类型转为json_string类型"""
    return json_format.MessageToJson(pb)


def json_to_pb_talk(json_obj):
    json_str = json.dumps(json_obj, indent=4)
    yield json_format.Parse(json_str, talk_pb2.TalkRequest())


def run(url, payload):
    with grpc.insecure_channel(url) as channel:
        stub = talk_pb2_grpc.TalkStub(channel)
        responses = stub.StreamingTalk(json_to_pb_talk(payload))
        response = list(responses)

    for r in response:
        r = json.loads(pb_to_json(r))
        print(r)


if __name__ == '__main__':

    run(url="172.16.23.85:30811", payload={
        "isfull": True,
        "testMode": False,
        "agentid": 666,
        "sessionid": str(uuid.uuid4()),
        "questionid": str(uuid.uuid4()),
        "eventtype": 0,
        "robotid": "123",
        "tenantcode": "cloudminds",
        "version": "v3",
        "asr": {
            "lang": "CH",
            "text": "现在几点了"
        }
    })
