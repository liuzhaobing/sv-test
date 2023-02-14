# -*- coding:utf-8 -*-
import uuid

import json
import grpc
import tts_pb2
import tts_pb2_grpc

from google.protobuf import json_format


def pb_to_json(pb):
    """将google.protobuf.struct类型转为json_string类型"""
    return json_format.MessageToJson(pb)


def json_to_pb_tts_call(json_obj):
    """https://blog.csdn.net/hsy12342611/article/details/128108829"""
    json_str = json.dumps(json_obj, indent=4)
    return json_format.Parse(json_str, tts_pb2.TtsReq())


def run(url, payload):
    with grpc.insecure_channel(url) as channel:
        stub = tts_pb2_grpc.CloudMindsTTSStub(channel)
        responses = stub.Call(json_to_pb_tts_call(payload))
        response = list(responses)

    for r in response:
        r = json.loads(pb_to_json(r))
        print(r)
    print(len(response))


if __name__ == '__main__':
    run(url="172.16.23.15:31349", payload={
        "text": "《登高》，唐，杜甫，风急天高猿啸哀，渚清沙白鸟飞回。无边落木萧萧下，不尽长江滚滚来。万里悲秋常作客，百年多病独登台。艰难苦恨繁霜鬓，潦倒新停浊酒杯。",
        "speed": "3",
        "volume": "3",
        "pitch": "medium",
        "emotions": "Gentle",
        "parameter_speaker_name": "DaSiRu",
        "parameter_digital_person": "SweetGirl",
        "parameter_flag": {
            "mouth": "true",
            "movement": "true",
            "expression": "true"
        }
    })
