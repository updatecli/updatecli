#!/usr/bin/env python3
"""Summarizes the OpenTelemetry spans an e2e run exported, as markdown.

usage: trace_summary.py SPANS_JSONL [--json OUT]

SPANS_JSONL is the OTLP/JSON file a collector's file exporter writes, one export per line.
--json also writes the aggregates, for comparing runs.
"""

import argparse
import collections
import json
from urllib.parse import urlparse

STAGES = [
    "updatecli.prepare",
    "updatecli.load_configurations",
    "updatecli.init_scm",
    "updatecli.autodiscovery",
    "updatecli.run",
    "updatecli.push_commits",
    "updatecli.run_actions",
    "updatecli.prune_scm_branches",
]


def load(path: str) -> list[dict]:
    spans = []
    with open(path) as f:
        for line in f:
            for resource in json.loads(line)["resourceSpans"]:
                for scope in resource["scopeSpans"]:
                    for span in scope["spans"]:
                        span["attrs"] = {a["key"]: next(iter(a["value"].values()), None) for a in span.get("attributes", [])}
                        span["seconds"] = (int(span["endTimeUnixNano"]) - int(span["startTimeUnixNano"])) / 1e9
                        spans.append(span)
    return spans


def aggregate(spans: list[dict]) -> dict:
    def group(selected, key):
        groups = collections.defaultdict(list)
        for span in selected:
            groups[key(span)].append(span["seconds"])
        return {
            name: {"count": len(v), "total_s": round(sum(v), 1), "avg_s": round(sum(v) / len(v), 3), "max_s": round(max(v), 1)}
            for name, v in sorted(groups.items(), key=lambda item: -sum(item[1]))
        }

    def http_key(span):
        url = urlparse(str(span["attrs"].get("url.full", "")))
        return f"{span['attrs'].get('http.request.method', '?')} {url.netloc}{'/graphql' if url.path.endswith('/graphql') else ''}"

    roots = [s for s in spans if s["name"] == "updatecli"]
    pipelines = [s for s in spans if s["name"] == "updatecli.pipeline"]
    return {
        "processes": len(roots),
        "wall_s": round(sum(s["seconds"] for s in roots), 1),
        "pipelines": len(pipelines),
        "pipeline_results": dict(collections.Counter(str(s["attrs"].get("updatecli.pipeline.result")) for s in pipelines)),
        "stages": group([s for s in spans if s["name"] in STAGES], lambda s: s["name"]),
        "resources": group(
            [s for s in spans if s["name"] == "updatecli.resource"],
            lambda s: f"{s['attrs'].get('updatecli.resource.category')}/{s['attrs'].get('updatecli.resource.kind')}",
        ),
        "http": group([s for s in spans if s["name"].startswith("HTTP ")], http_key),
        "slowest_pipelines": [
            {"name": s["attrs"].get("updatecli.pipeline.name"), "result": s["attrs"].get("updatecli.pipeline.result"), "seconds": round(s["seconds"], 1)}
            for s in sorted(pipelines, key=lambda s: -s["seconds"])[:10]
        ],
    }


def table(title: str, groups: dict, limit: int = 15) -> str:
    rows = [f"### {title}", "", "| | count | total s | avg s | max s |", "|---|---:|---:|---:|---:|"]
    rows += [f"| `{name}` | {g['count']} | {g['total_s']} | {g['avg_s']} | {g['max_s']} |" for name, g in list(groups.items())[:limit]]
    return "\n".join(rows) + "\n"


def markdown(summary: dict) -> str:
    out = [
        "## Updatecli e2e trace summary",
        "",
        f"{summary['processes']} updatecli process(es), {summary['wall_s']}s in total, {summary['pipelines']} pipelines "
        f"({', '.join(f'{k}: {v}' for k, v in sorted(summary['pipeline_results'].items()))}).",
        "",
        "Resources of a pipeline run concurrently, so resource totals can exceed the wall time.",
        "",
        table("Stages", summary["stages"]),
        table("Resources by category/kind", summary["resources"]),
        table("HTTP requests by method and host", summary["http"]),
        "### Slowest pipelines",
        "",
        "| pipeline | result | s |",
        "|---|---|---:|",
    ]
    out += [f"| {p['name']} | {p['result']} | {p['seconds']} |" for p in summary["slowest_pipelines"]]
    return "\n".join(out) + "\n"


def main() -> None:
    parser = argparse.ArgumentParser(description="Summarizes the OpenTelemetry spans an e2e run exported.")
    parser.add_argument("spans")
    parser.add_argument("--json")
    args = parser.parse_args()

    summary = aggregate(load(args.spans))
    if args.json:
        with open(args.json, "w") as f:
            json.dump(summary, f, indent=2)
    print(markdown(summary), end="")


if __name__ == "__main__":
    main()
