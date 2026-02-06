export { RoomDurableObject } from './room';

interface Env {
  ROOM: DurableObjectNamespace;
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url);
    const upgrade = request.headers.get('Upgrade');

    // WebSocket upgrade
    if (upgrade?.toLowerCase() === 'websocket') {
      const roomKey = url.searchParams.get('roomKey');
      if (!roomKey) {
        return new Response('Missing roomKey query parameter', { status: 400 });
      }
      const id = env.ROOM.idFromName(roomKey);
      const stub = env.ROOM.get(id);
      return stub.fetch(request);
    }

    // Health check
    if (url.pathname === '/' && request.method === 'GET') {
      return new Response(JSON.stringify({ status: 'ok' }), {
        headers: { 'Content-Type': 'application/json' },
      });
    }

    return new Response('WebSocket upgrade required', { status: 426 });
  },
};
