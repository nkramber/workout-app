import { Code, ConnectError } from "@connectrpc/connect";

// online is false only when the browser knows that it has no network.
function online(): boolean {
  return typeof navigator === "undefined" || navigator.onLine !== false;
}

// isNoConnection tells whether a call failed because the device has no
// connection to the API. A fetch that can not reach the server fails with
// a TypeError, and Connect gives it the code `unknown` with the TypeError
// as its cause.
export function isNoConnection(err: unknown, isOnline = online()): boolean {
  const e = ConnectError.from(err);
  if (!isOnline || e.code === Code.Unavailable) return true;
  return e.code === Code.Unknown && e.cause instanceof TypeError;
}

// SERVER_FAULT is the message of the code `internal`. The API answered,
// but it failed, so the text does not say that the API did not answer
// (D-206).
const SERVER_FAULT = "The server failed.";

// changeErrorText gives the message of a failed change of the inventory
// or of the profile.
// Phase 4 has no outbox, so a change with no connection is not saved, and
// the owner tries again (D-196). The text holds no value of the request.
export function changeErrorText(err: unknown, isOnline = online()): string {
  if (isNoConnection(err, isOnline)) return "No connection. The change is not saved. Try again when the phone is online.";
  switch (ConnectError.from(err).code) {
    case Code.InvalidArgument:
      return "The server did not accept the values. Read them again.";
    case Code.FailedPrecondition:
      return "The weights changed after this screen showed them. Read the new weights, then confirm again.";
    case Code.NotFound:
      return "The inventory does not hold this item now.";
    case Code.PermissionDenied:
      return "This account is not on the allowlist.";
    case Code.Unauthenticated:
      return "The API did not accept the sign-in.";
    case Code.Internal:
      return SERVER_FAULT + " The change is not saved.";
    default:
      return "The API did not answer. The change is not saved.";
  }
}

// loadErrorText gives the message of a failed read of the catalog, the
// inventory, the profile, or the lists of the profile.
export function loadErrorText(err: unknown, isOnline = online()): string {
  if (isNoConnection(err, isOnline)) return "No connection. Try again when the phone is online.";
  switch (ConnectError.from(err).code) {
    case Code.Internal:
      return SERVER_FAULT;
    case Code.PermissionDenied:
      return "This account is not on the allowlist.";
    case Code.Unauthenticated:
      return "The API did not accept the sign-in.";
    default:
      return "The API did not answer.";
  }
}
