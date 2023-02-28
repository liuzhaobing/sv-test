# Generated by the gRPC Python protocol compiler plugin. DO NOT EDIT!
"""Client and server classes corresponding to protobuf-defined services."""
import grpc

from google.protobuf import empty_pb2 as google_dot_protobuf_dot_empty__pb2
import ner_pb2 as ner__pb2


class CloudMindsNerStub(object):
    """Missing associated documentation comment in .proto file."""

    def __init__(self, channel):
        """Constructor.

        Args:
            channel: A grpc.Channel.
        """
        self.NerCall = channel.unary_unary(
                '/api.ner.v1.CloudMindsNer/NerCall',
                request_serializer=ner__pb2.NerReq.SerializeToString,
                response_deserializer=ner__pb2.NerRes.FromString,
                )
        self.GetVersion = channel.unary_unary(
                '/api.ner.v1.CloudMindsNer/GetVersion',
                request_serializer=google_dot_protobuf_dot_empty__pb2.Empty.SerializeToString,
                response_deserializer=ner__pb2.VerRsp.FromString,
                )


class CloudMindsNerServicer(object):
    """Missing associated documentation comment in .proto file."""

    def NerCall(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def GetVersion(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')


def add_CloudMindsNerServicer_to_server(servicer, server):
    rpc_method_handlers = {
            'NerCall': grpc.unary_unary_rpc_method_handler(
                    servicer.NerCall,
                    request_deserializer=ner__pb2.NerReq.FromString,
                    response_serializer=ner__pb2.NerRes.SerializeToString,
            ),
            'GetVersion': grpc.unary_unary_rpc_method_handler(
                    servicer.GetVersion,
                    request_deserializer=google_dot_protobuf_dot_empty__pb2.Empty.FromString,
                    response_serializer=ner__pb2.VerRsp.SerializeToString,
            ),
    }
    generic_handler = grpc.method_handlers_generic_handler(
            'api.ner.v1.CloudMindsNer', rpc_method_handlers)
    server.add_generic_rpc_handlers((generic_handler,))


 # This class is part of an EXPERIMENTAL API.
class CloudMindsNer(object):
    """Missing associated documentation comment in .proto file."""

    @staticmethod
    def NerCall(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/api.ner.v1.CloudMindsNer/NerCall',
            ner__pb2.NerReq.SerializeToString,
            ner__pb2.NerRes.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def GetVersion(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/api.ner.v1.CloudMindsNer/GetVersion',
            google_dot_protobuf_dot_empty__pb2.Empty.SerializeToString,
            ner__pb2.VerRsp.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)
