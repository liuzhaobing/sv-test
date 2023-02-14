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


def run():
    def talk_req():
        yield talk_pb2.TalkRequest(is_full=True,
                                   agent_id=666,
                                   session_id=str(uuid.uuid4()),
                                   question_id=str(uuid.uuid4()),
                                   event_type=0,
                                   robot_id="123",
                                   tenant_code="cloudminds",
                                   version="v3",
                                   test_mode=False,
                                   asr=talk_pb2.Asr(lang="CH", text="背一首杜甫的诗"))

    with grpc.insecure_channel('172.16.23.85:30811') as channel:
        stub = talk_pb2_grpc.TalkStub(channel)
        responses = stub.StreamingTalk(talk_req().__iter__())
        response = list(responses)

    for r in response:
        r = json.loads(pb_to_json(r))
        print(r)


if __name__ == '__main__':
    run()
