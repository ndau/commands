#!/usr/bin/env python3
import csv, json, sys

if __name__ == "__main__":
    input = open(sys.argv[1])
    data = json.load(input)
    input.close()

    output = csv.writer(sys.stdout)

    output.writerow(data[0].keys())  # header row

    for row in data:
        output.writerow(row.values())

