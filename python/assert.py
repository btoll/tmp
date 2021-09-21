import json
import os
import subprocess

import requests


def main():
    # Variable names taken from RFC: https://datatracker.ietf.org/doc/html/rfc6763#section-4.1
    instance = os.environ["PD_SERVICE"]
    if not instance:
        raise RuntimeError("No instance name")

    environment = "stg"
    # ex. stg.event-storage-ex-api.service.consul
    instance_name = ".".join((environment, instance, "service", "consul"))

    print("-------------------------------")
    print("-------------------------------")
    print("[INFO] Instance name: ", instance_name)

    port = subprocess.getoutput(f"dig +short {instance_name} SRV | cut -d\  -f3")
    if not port:
        raise RuntimeError("No service port")

    socket = ":".join((instance_name, port))
    service = "://".join(("http", socket))

    test_dir = os.environ["TEST_DIR"]
    test_file = "assert.json"
    test_path = "/".join(("/workdir", test_dir, test_file))

    print("-------------------------------")
    print(f"[INFO] Loading tests in {test_path}")

    with open(test_path, "r") as fp:
        j = json.load(fp)

    print("-------------------------------")
    print("[INFO] Service: ", service)
    print("-------------------------------")

    passed = 0
    for test in j["tests"]:
        response = requests.request(
            test["action"] if "action" in test else "GET",
            "/".join((service, test["endpoint"])),
            headers = {
                "Content-Type": "application/json"
            },
            json = test["body"] if "body" in test else "",
        )

        try:
            assert response.status_code == (test["status_code"] if "status_code" in test else 200), "Status code does not indicate a successful request"
            assert json.loads(response.text) == test["assert"], "Response text does not equal the expected text"
            print(f"{test['name']}...passed")
            passed += 1
        except AssertionError:
            print(f"{test['name']}...failed")

    total_tests = len(j["tests"])
    print("-------------------------------")
    print("[INFO] Test results:")
    print(f"Total tests: {total_tests}")
    print(f"\t\t{passed} passed")
    print(f"\t\t{total_tests - passed} failed")

    if passed != total_tests:
        raise Exception("One or more tests failed, exiting...")

    print("-------------------------------")
    print("-------------------------------")


if __name__ == "__main__":
    main()
