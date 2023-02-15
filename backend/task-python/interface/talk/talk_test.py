# -*- coding:utf-8 -*-
import uuid

import json
import talk_pb2
import talk_pb2_grpc
from interface.interface import Interface, pb_to_json

if __name__ == '__main__':
    payload = {"isfull": True,
               "testMode": False,
               "agentid": 666,
               "sessionid": str(uuid.uuid4()),
               "questionid": str(uuid.uuid4()),
               "eventtype": 0,
               "robotid": "123",
               "version": "v3",
               "tenantcode": "cloudminds",
               "asr": {"lang": "CH", "text": "背一首杜甫的登高"}
               }
    ins = Interface(url="172.16.23.85:30811", stub=talk_pb2_grpc.TalkStub)
    result = ins.call(message=talk_pb2.TalkRequest(), func=ins.stub.StreamingTalk, payload=payload, iterator=True)
    response_json = [json.loads(pb_to_json(response)) for response in result]
    print(response_json)
