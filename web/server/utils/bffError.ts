interface SafeBffError {
  error: true;
  statusCode: number;
  statusMessage: string;
  message: string;
}

const genericMessages: Record<number, string> = {
  500: "BFF request failed",
  502: "Downstream service is unavailable",
  503: "Downstream service is unavailable",
  504: "Downstream service timed out",
};

export function sanitizeBffError(error: unknown): SafeBffError {
  const candidate =
    error && typeof error === "object"
      ? (error as { statusCode?: unknown; statusMessage?: unknown })
      : {};
  const rawStatus = Number(candidate.statusCode);
  const statusCode =
    Number.isInteger(rawStatus) && rawStatus >= 400 && rawStatus <= 599
      ? rawStatus
      : 500;
  const candidateMessage =
    typeof candidate.statusMessage === "string"
      ? candidate.statusMessage.trim()
      : "";
  const statusMessage =
    genericMessages[statusCode] ||
    (candidateMessage &&
    candidateMessage.length <= 120 &&
    !/[\r\n]/u.test(candidateMessage)
      ? candidateMessage
      : "BFF request was rejected");
  return {
    error: true,
    statusCode,
    statusMessage,
    message: statusMessage,
  };
}
