# -*- coding:utf-8 -*-
import wave
import tts_pb2
import tts_pb2_grpc
from interface.interface import Interface


def write_wav(pcm_bytes, wav_file):
    with wave.open(wav_file, "wb") as w:
        w.setnchannels(1)
        w.setsampwidth(2)
        w.setframerate(16000)
        w.writeframes(pcm_bytes)


if __name__ == '__main__':
    payload = {
        "text": "《登高》，唐，杜甫，风急天高猿啸哀，渚清沙白鸟飞回。无边落木萧萧下，不尽长江滚滚来。万里悲秋常作客，百年多病独登台。艰难苦恨繁霜鬓，潦倒新停浊酒杯。",
        "speed": "3",
        "volume": "3",
        "pitch": "medium",
        "emotions": "Gentle",
        "parameter_speaker_name": "DaXiaoFang",
        "parameter_digital_person": "SweetGirl",
        "parameter_flag": {
            "mouth": "true",
            "movement": "true",
            "expression": "true"
        }
    }
    url = "172.16.23.15:31349"  # fit-86
    # url = "172.16.23.17:31795"  # sit-134
    ins = Interface(url=url, stub=tts_pb2_grpc.CloudMindsTTSStub)
    result = ins.call(message=tts_pb2.TtsReq(), func=ins.stub.Call, payload=payload)
    pcm = b""
    for r in result:
        pcm += r.synthesized_audio.pcm
    write_wav(pcm, f"./test.wav")
