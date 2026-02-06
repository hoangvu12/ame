import { DurableObject } from 'cloudflare:workers';

interface Env {
  ROOM: DurableObjectNamespace;
}

interface SkinInfo {
  championId: string;
  skinId: string;
  baseSkinId: string;
  championName: string;
  skinName: string;
  chromaName: string;
}

interface MemberAttachment {
  puuid: string;
  skinInfo: SkinInfo;
}

interface JoinMessage {
  type: 'join';
  puuid: string;
  skinInfo: SkinInfo;
}

interface SkinMessage {
  type: 'skin';
  skinInfo: SkinInfo;
}

interface LeaveMessage {
  type: 'leave';
}

type ClientMessage = JoinMessage | SkinMessage | LeaveMessage;

export class RoomDurableObject extends DurableObject<Env> {
  constructor(ctx: DurableObjectState, env: Env) {
    super(ctx, env);
    this.ctx.setWebSocketAutoResponse(
      new WebSocketRequestResponsePair('ping', 'pong'),
    );
  }

  async fetch(request: Request): Promise<Response> {
    const pair = new WebSocketPair();
    const [client, server] = Object.values(pair);
    this.ctx.acceptWebSocket(server);
    return new Response(null, { status: 101, webSocket: client });
  }

  async webSocketMessage(ws: WebSocket, message: string | ArrayBuffer) {
    if (typeof message !== 'string') return;
    if (message === 'pong') return;

    let msg: ClientMessage;
    try {
      msg = JSON.parse(message);
    } catch {
      return;
    }

    switch (msg.type) {
      case 'join': {
        const attachment: MemberAttachment = {
          puuid: msg.puuid,
          skinInfo: msg.skinInfo || ({} as SkinInfo),
        };
        ws.serializeAttachment(attachment);
        this.broadcastMembers();
        break;
      }
      case 'skin': {
        const existing = ws.deserializeAttachment() as MemberAttachment | null;
        if (existing) {
          existing.skinInfo = msg.skinInfo || ({} as SkinInfo);
          ws.serializeAttachment(existing);
          this.broadcastMembers();
        }
        break;
      }
      case 'leave': {
        ws.close(1000, 'client left');
        break;
      }
    }
  }

  async webSocketClose() {
    this.broadcastMembers();
  }

  async webSocketError() {
    this.broadcastMembers();
  }

  private broadcastMembers() {
    const sockets = this.ctx.getWebSockets();
    const members: MemberAttachment[] = [];

    for (const ws of sockets) {
      try {
        if (ws.readyState !== WebSocket.READY_STATE_OPEN) continue;
        const att = ws.deserializeAttachment() as MemberAttachment | null;
        if (att?.puuid) {
          members.push(att);
        }
      } catch {
        // skip broken sockets
      }
    }

    const payload = JSON.stringify({ type: 'members', members });
    for (const ws of sockets) {
      try {
        if (ws.readyState === WebSocket.READY_STATE_OPEN) {
          ws.send(payload);
        }
      } catch {
        // skip
      }
    }
  }
}
