import { connect } from 'cloudflare:sockets'

interface Env {
    Authorization?: string
}

export default {
    async fetch(request: Request, env: Env, ctx: ExecutionContext): Promise<Response> {
        if (request.cf?.httpProtocol !== 'HTTP/2') {
            return new Response(null, { status: 400 })
        }
        if (!request.headers.get('Content-Type')?.startsWith('application/grpc')) {
            return new Response(null, { status: 400 })
        }
        if (env.Authorization && request.headers.get('Proxy-Authorization') !== 'basic ' + btoa(env.Authorization)) {
            return new Response(null, { status: 403 })
        }

        const url = new URL(request.url)
        const target = url.searchParams.get("target")
        if (!target) {
            return new Response(null, { status: 404 })
        } else if (target.endsWith('25')) {
            return new Response('Connections to port 25 are prohibited', { status: 400 })
        }

        return http2relay(request.body as ReadableStream<Uint8Array>, target)
    }
}

async function http2relay(body: ReadableStream, target: string): Promise<Response> {
    const tcpSocket = connect(target)
    const start = performance.now()
    return tcpSocket.opened
        .then(() => {
            const end = performance.now()
            console.log({ target, "TCP handshake": end - start })
            body.pipeTo(tcpSocket.writable)
            return new Response(tcpSocket.readable.pipeThrough(new TransformStream()))
        })
        .catch(reason => {
            const err = reason as Error
            if (err)
                console.error(`TCP connect to ${target} fail`, err.message)
            return new Response(err.message, {
                status: 400
            })
        })
}