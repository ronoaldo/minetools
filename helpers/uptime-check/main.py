#!/usr/bin/env python

import socket
import time
import os
from flask import Flask, request

app = Flask(__name__)

# Returns ping time in seconds (up), False (down), or None (error).
# Borrowed from https://github.com/minetest/serverlist/blob/5d5f31d295d8ed3b94a70a4470e37481facd6f9d/server.py#L140
def serverUp(info):
    try:
        sock = socket.socket(info[0], info[1], info[2])
        sock.settimeout(3)
        sock.connect(info[4])
        # send packet of type ORIGINAL, with no data
        #     this should prompt the server to assign us a peer id
        # [0] u32       protocol_id (PROTOCOL_ID)
        # [4] session_t sender_peer_id (PEER_ID_INEXISTENT)
        # [6] u8        channel
        # [7] u8        type (PACKET_TYPE_ORIGINAL)
        buf = b"\x4f\x45\x74\x03\x00\x00\x00\x01"
        sock.send(buf)
        start = time.time()
        # receive reliable packet of type CONTROL, subtype SET_PEER_ID,
        #     with our assigned peer id as data
        # [0] u32        protocol_id (PROTOCOL_ID)
        # [4] session_t  sender_peer_id
        # [6] u8         channel
        # [7] u8         type (PACKET_TYPE_RELIABLE)
        # [8] u16        seqnum
        # [10] u8        type (PACKET_TYPE_CONTROL)
        # [11] u8        controltype (CONTROLTYPE_SET_PEER_ID)
        # [12] session_t peer_id_new
        data = sock.recv(1024)
        end = time.time()
        if not data:
            return False
        peer_id = data[12:14]
        # send packet of type CONTROL, subtype DISCO,
        #     to cleanly close our server connection
        # [0] u32       protocol_id (PROTOCOL_ID)
        # [4] session_t sender_peer_id
        # [6] u8        channel
        # [7] u8        type (PACKET_TYPE_CONTROL)
        # [8] u8        controltype (CONTROLTYPE_DISCO)
        buf = b"\x4f\x45\x74\x03" + peer_id + b"\x00\x00\x03"
        sock.send(buf)
        sock.close()
        return end - start
    except socket.timeout:
        return False
    except Exception as e:
        print(e)
        return None

@app.route("/status")
def status():
    host = request.args.get("host")
    port = request.args.get("port")
    info = socket.getaddrinfo(host, port, type=socket.SOCK_DGRAM, proto=socket.SOL_UDP)
    ping = serverUp(info[0])
    if not ping:
        return "Server unavailable", 402
    return "Server up. Ping %.04fs" % ping

if __name__ == "__main__":
    app.run("0.0.0.0", os.getenv("PORT"))
