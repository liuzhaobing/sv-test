# -*- coding:utf-8 -*-
import uuid
import wave

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


def write_wav(pcm_bytes, wav_file):
    with wave.open(wav_file, "wb") as w:
        w.setnchannels(1)
        w.setsampwidth(2)
        w.setframerate(16000)
        w.writeframes(pcm_bytes)


def run(url, payload):
    with grpc.insecure_channel(url) as channel:
        stub = tts_pb2_grpc.CloudMindsTTSStub(channel)
        responses = stub.Call(json_to_pb_tts_call(payload))
        response = list(responses)

    pcm = b""
    config_text = None
    debug_info = []

    for r in response:
        pcm += r.synthesized_audio.pcm
        r = json.loads(pb_to_json(r))
        config_text = r["configText"] if r.__contains__("configText") else config_text
        debug_info.append(r["debugInfo"]) if r.__contains__("debugInfo") else None

    write_wav(pcm, f"./test.wav")


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
