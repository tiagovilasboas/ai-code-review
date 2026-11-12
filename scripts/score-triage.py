"""Score a triage run against a labelled blind set.

usage: python3 -I scripts/score-triage.py <review-binary> <blind-dir> <verdicts-dir>

For each case in <blind-dir>, the engine reviews change.diff; when
<verdicts-dir>/<case>.json exists, it reviews again with --triage. Each
original finding is a TP when expected.json lists its path, line and class,
else an FP. The script counts the TPs and FPs the verdicts dropped, or
downgraded from BLOCK (HIGH or CRITICAL) to a lower severity.
"""
import json, os, subprocess, sys

BLOCK = {"HIGH", "CRITICAL"}


def review(engine, diff, verdicts=None):
    cmd = [engine, "--format", "json"]
    if verdicts:
        cmd += ["--triage", verdicts]
    p = subprocess.run(cmd + [diff], capture_output=True, text=True)
    if p.returncode == 2:
        raise SystemExit(f"{diff}: engine rejected the run: {p.stderr.strip()}")
    return json.loads(p.stdout)


def main(engine, blind, verdicts_dir):
    stats = dict.fromkeys(["tp", "fp", "tp_dropped", "tp_downgraded",
                           "fp_dropped", "fp_downgraded"], 0)
    for case in sorted(os.listdir(blind)):
        diff = os.path.join(blind, case, "change.diff")
        if not os.path.isfile(diff):
            continue
        exp = json.load(open(os.path.join(blind, case, "expected.json")))
        want = {(f["path"], f["line"], f["class"]) for f in exp.get("findings", [])}
        before = review(engine, diff)["findings"]
        vfile = os.path.join(verdicts_dir, case + ".json")
        after = review(engine, diff, vfile)["findings"] if os.path.exists(vfile) else before
        post = {f["id"]: f for f in after}
        for f in before:
            tag = "tp" if (f["path"], f["line"], f["findingClass"]) in want else "fp"
            stats[tag] += 1
            g = post.get(f["id"])
            if g is None:
                stats[tag + "_dropped"] += 1
                print(f"{case}\t{tag.upper()} dropped\t{f['findingClass']}")
            elif f["severity"] in BLOCK and g["severity"] not in BLOCK:
                stats[tag + "_downgraded"] += 1
                print(f"{case}\t{tag.upper()} downgraded\t{f['findingClass']}")
    print(json.dumps(stats))


if __name__ == "__main__":
    if len(sys.argv) != 4:
        raise SystemExit(__doc__)
    main(*sys.argv[1:])
