import argparse
import json
import random
import time
import urllib.error
import urllib.request
from datetime import datetime, timezone

FIRST_NAMES = [
    "Aarav", "Diya", "Ishaan", "Meera", "Rohan", "Kavya", "Arjun", "Nisha"
]
LAST_NAMES = ["Kumar", "Reddy", "Sharma", "Patel", "Iyer", "Verma", "Gupta", "Rao"]
AREAS = [
    "Tambaram", "Selaiyur", "Pallikaranai", "Chromepet", "Velachery", "Medavakkam"
]
PAYMENT_METHODS = ["upi", "cod", "card"]
PACKAGE_SIZES = ["small", "medium", "large"]


def now_iso() -> str:
    return datetime.now(timezone.utc).isoformat().replace("+00:00", "Z")


def random_order(idx: int) -> dict:
    first = random.choice(FIRST_NAMES)
    last = random.choice(LAST_NAMES)
    customer_name = f"{first} {last}"
    payment_method = random.choice(PAYMENT_METHODS)
    amount = round(random.uniform(120, 1450), 2)
    package_size = random.choice(PACKAGE_SIZES)
    weight = {
        "small": round(random.uniform(0.2, 2.0), 2),
        "medium": round(random.uniform(2.1, 7.0), 2),
        "large": round(random.uniform(7.1, 20.0), 2),
    }[package_size]

    return {
        "order_id": f"ord-fake-{int(time.time())}-{idx}",
        "customer_name": customer_name,
        "customer_id": f"cust-{random.randint(10000, 99999)}",
        "customer_contact_number": f"+91{random.randint(9000000000, 9999999999)}",
        "address": f"{random.randint(1, 400)}, {random.choice(AREAS)}, Chennai",
        "payment_method": payment_method,
        "payment_amount": amount,
        "package_size": package_size,
        "package_weight_kg": weight,
        "status": "assigned",
        "assigned_at": now_iso(),
        "source": "fake-order-publisher",
    }


def post_json(url: str, payload: dict, timeout: int) -> tuple[int, str]:
    body = json.dumps(payload).encode("utf-8")
    req = urllib.request.Request(
        url,
        data=body,
        headers={"Content-Type": "application/json"},
        method="POST",
    )
    try:
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            text = resp.read().decode("utf-8", errors="replace")
            return resp.getcode(), text
    except urllib.error.HTTPError as http_err:
        text = http_err.read().decode("utf-8", errors="replace")
        return http_err.code, text


def request_json(method: str, url: str, timeout: int, payload: dict | None = None, headers: dict | None = None) -> tuple[int, str]:
    data = None
    req_headers = {"Content-Type": "application/json"}
    if headers:
        req_headers.update(headers)
    if payload is not None:
        data = json.dumps(payload).encode("utf-8")

    req = urllib.request.Request(url, data=data, headers=req_headers, method=method)
    try:
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            text = resp.read().decode("utf-8", errors="replace")
            return resp.getcode(), text
    except urllib.error.HTTPError as http_err:
        text = http_err.read().decode("utf-8", errors="replace")
        return http_err.code, text


def get_publisher_status(status_url: str, timeout: int, control_key: str) -> dict | None:
    headers = {}
    if control_key:
        headers["X-Publisher-Key"] = control_key
    status, text = request_json("GET", status_url, timeout, headers=headers)
    if status >= 400:
        print(f"[control] status check failed status={status} response={text}")
        return None
    try:
        return json.loads(text)
    except json.JSONDecodeError:
        print(f"[control] invalid status payload: {text}")
        return None


def main() -> None:
    parser = argparse.ArgumentParser(
        description="Send fake orders to worker backend at a fixed interval"
    )
    parser.add_argument(
        "--url",
        default="http://192.168.1.6:8001/api/v1/demo/orders/ingest",
        help="Ingest endpoint URL",
    )
    parser.add_argument(
        "--status-url",
        default="http://192.168.1.6:8001/api/v1/demo/orders/publisher/status",
        help="Backend control status URL (publishes only when active=true)",
    )
    parser.add_argument(
        "--control-key",
        default="",
        help="Optional X-Publisher-Key header value for control endpoints",
    )
    parser.add_argument(
        "--poll-seconds",
        type=int,
        default=5,
        help="How often to poll backend control status while waiting",
    )
    parser.add_argument(
        "--runs",
        type=int,
        default=2,
        help="How many publish cycles (default: 2)",
    )
    parser.add_argument(
        "--continuous",
        action="store_true",
        help="Run forever until interrupted (Ctrl+C)",
    )
    parser.add_argument(
        "--interval-seconds",
        type=int,
        default=60,
        help="Seconds between cycles (default: 60 => twice in 2 minutes)",
    )
    parser.add_argument(
        "--orders-per-run",
        type=int,
        default=1,
        help="Orders published per cycle",
    )
    parser.add_argument(
        "--timeout-seconds",
        type=int,
        default=15,
        help="HTTP timeout in seconds",
    )
    args = parser.parse_args()

    print(f"Publishing to: {args.url}")
    print(f"Control status URL: {args.status_url}")
    if args.continuous:
        print(
            f"Plan: continuous mode, {args.orders_per_run} order(s)/cycle, every {args.interval_seconds}s"
        )
        print(
            f"Behavior: publish only when backend initiates; lease expires after 5 minutes unless ack extends it"
        )
    else:
        print(
            f"Plan: {args.runs} runs, {args.orders_per_run} order(s)/run, every {args.interval_seconds}s"
        )

    run_no = 0
    current_session = ""
    try:
        while True:
            status_payload = get_publisher_status(args.status_url, args.timeout_seconds, args.control_key)
            if status_payload is None:
                time.sleep(max(1, args.poll_seconds))
                continue

            active = bool(status_payload.get("active", False))
            session_id = str(status_payload.get("session_id", "") or "")
            remaining_sec = int(status_payload.get("remaining_sec", 0) or 0)

            if not active:
                if current_session:
                    print(f"[control] lease expired for session {current_session}. Waiting for backend initiate/ack.")
                    current_session = ""
                time.sleep(max(1, args.poll_seconds))
                continue

            if session_id and session_id != current_session:
                current_session = session_id
                print(f"[control] active lease session={current_session}, remaining={remaining_sec}s")

            run_no += 1
            if args.continuous:
                print(f"\nRun {run_no} at {now_iso()} (session={current_session}, lease_remaining={remaining_sec}s)")
            else:
                print(f"\\nRun {run_no}/{args.runs} at {now_iso()}")

            for i in range(args.orders_per_run):
                payload = random_order(i + 1)
                status, response_text = post_json(args.url, payload, args.timeout_seconds)
                print(f"  -> {payload['order_id']} status={status}")
                if status >= 400:
                    print(f"     response={response_text}")

            if not args.continuous and run_no >= args.runs:
                break

            time.sleep(args.interval_seconds)
    except KeyboardInterrupt:
        print("\\nStopped by user.")

    print("\\nDone.")


if __name__ == "__main__":
    main()
