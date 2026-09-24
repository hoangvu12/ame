const MAX_LOGS = 500;

const logs = [];

function formatTimestamp(date) {
  return date.toISOString().replace('T', ' ').slice(0, 19);
}

function addLog(level, tag, message, ...args) {
	// Keep only launch/apply and recovery diagnostics; omit UI/selection chatter.
	if (!((tag === 'ws' && /WebSocket|Unstuck|apply sent|onmessage error/i.test(message)) ||
	    (tag === 'auto' && /applying|sending as apply|apply in-flight/i.test(message)))) return;
  const timestamp = new Date();
  const formattedArgs = args.map(arg => {
    if (typeof arg === 'object') {
      try {
        return JSON.stringify(arg);
      } catch {
        return String(arg);
      }
    }
    return String(arg);
  }).join(' ');

  const fullMessage = formattedArgs ? `${message} ${formattedArgs}` : message;

  logs.push({
    timestamp,
    level,
    tag,
    message: fullMessage,
  });

  if (logs.length > MAX_LOGS) {
    logs.shift();
  }

  // Also log to actual console
  const prefix = `[ame:${tag}]`;
  const consoleMsg = `${prefix} ${fullMessage}`;

  switch (level) {
    case 'error':
      console.error(consoleMsg);
      break;
    case 'warn':
      console.warn(consoleMsg);
      break;
    default:
      console.log(consoleMsg);
  }
}

export function createLogger(tag) {
  return {
    log: (message, ...args) => addLog('info', tag, message, ...args),
    info: (message, ...args) => addLog('info', tag, message, ...args),
    warn: (message, ...args) => addLog('warn', tag, message, ...args),
    error: (message, ...args) => addLog('error', tag, message, ...args),
  };
}

export function getPluginLogsJSON() {
  return logs.map(entry => ({
    timestamp: entry.timestamp.getTime(),
    source: `plugin:${entry.tag}`,
    level: entry.level,
    message: entry.message,
  }));
}

// Default logger for general use
export const log = createLogger('core');
