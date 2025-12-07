export type LogFn = (msg: string) => void;
type ErrorLike = Error | { message: string } | string | unknown;

export interface Logger {
    info: (msg: string) => void;
    warn: (msg: string) => void;
    error: (msg: string, err?: unknown) => void;
    ui: (msg: string) => void;
}

export function createLogger(uiCallback?: LogFn): Logger {
    function time() {
        return new Date().toISOString().split("T")[1].replace("Z", "");
    }

    return {
        info(msg) {
            console.log(`[${time()}] ${msg}`);
            uiCallback?.(msg);
        },
        warn(msg) {
            console.warn(`[WARN ${time()}] ${msg}`);
            uiCallback?.(msg);
        },
        error(msg, err) {
            console.error(`[ERROR ${time()}] ${msg}`, err ?? "");
            uiCallback?.(`${msg} ${extractMessage(err)}`);
        },
        ui(msg) {
            uiCallback?.(msg);
        }
    };
}

function extractMessage(err: ErrorLike): string {
    if (!err) return "";
    if (typeof err === "string") return err;
    if (err instanceof Error) return err.message;
    if (typeof err === "object" && "message" in err && typeof (err as any).message === "string") {
        return (err as any).message;
    }
    return JSON.stringify(err);
}