# -*- coding:utf-8 -*-
import traceback


def template(proto, url, payload, stub, req_func, call_func, iterator):
    return f"""
# -*- coding:utf-8 -*-
import uuid
import json
import interface.{proto}.{proto}_pb2 as pb2
import interface.{proto}.{proto}_pb2_grpc as pb2_grpc
from interface.interface import Interface


ins = Interface(url={url}, stub=pb2_grpc.{stub})
result = ins.call(message=pb2.{req_func}, func=ins.stub.{call_func}, payload={payload}, iterator={iterator})
proResult["result"] = result
    """


if __name__ == '__main__':
    proto_file = "talk"
    proResult = {"result": []}
    sc = template(proto=proto_file,
                  url="'172.16.23.85:30811'",
                  payload={"isfull": True,
                           "testMode": False,
                           "agentid": 666,
                           "sessionid": "123456",
                           "questionid": "123456",
                           "eventtype": 0,
                           "robotid": "123",
                           "version": "v3",
                           "tenantcode": "cloudminds",
                           "asr": {"lang": "CH", "text": "背一首杜甫的登高"}
                           },
                  stub="TalkStub",
                  req_func="TalkRequest()",
                  call_func="StreamingTalk",
                  iterator=True)
    try:
        exec(sc, {"proResult": proResult})
    except Exception as e:
        print(traceback.format_exc())

    print(proResult)
