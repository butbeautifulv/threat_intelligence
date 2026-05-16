#!/usr/bin/env python3
"""Download GAIA from Hugging Face (gated — requires HF token). Does not commit data."""
from __future__ import annotations

import argparse
import os
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[3]
DEFAULT_DIR = ROOT / "eval" / "gaia" / "data"


def main() -> int:
    ap = argparse.ArgumentParser(description="Snapshot gaia-benchmark/GAIA (gated dataset)")
    ap.add_argument("--data-dir", type=Path, default=Path(os.environ.get("GAIA_DATA_DIR", DEFAULT_DIR)))
    ap.add_argument("--config", default=os.environ.get("GAIA_CONFIG", "2023_level1"))
    ap.add_argument("--split", default=os.environ.get("GAIA_SPLIT", "validation"))
    args = ap.parse_args()

    token = os.environ.get("HF_TOKEN") or os.environ.get("HUGGING_FACE_HUB_TOKEN")
    if not token:
        print(
            "GAIA is gated on Hugging Face. Set HF_TOKEN and accept the dataset terms:\n"
            "  https://huggingface.co/datasets/gaia-benchmark/GAIA\n"
            "Do not commit downloaded files or publish validation/test answers.",
            file=sys.stderr,
        )
        return 2

    try:
        from huggingface_hub import snapshot_download
        from datasets import load_dataset
    except ImportError:
        print("pip install huggingface_hub datasets", file=sys.stderr)
        return 2

    args.data_dir.mkdir(parents=True, exist_ok=True)
    data_dir = snapshot_download(
        repo_id="gaia-benchmark/GAIA",
        repo_type="dataset",
        token=token,
        local_dir=str(args.data_dir / "hf-snapshot"),
    )
    ds = load_dataset(data_dir, args.config, split=args.split)
    out = args.data_dir / f"{args.config}-{args.split}.jsonl"
    with out.open("w", encoding="utf-8") as f:
        for row in ds:
            f.write(
                __import__("json").dumps(
                    {
                        "task_id": row.get("task_id"),
                        "Question": row.get("Question"),
                        "Level": row.get("Level"),
                        "Final answer": row.get("Final answer"),
                        "file_name": row.get("file_name", ""),
                        "file_path": row.get("file_path", ""),
                    },
                    ensure_ascii=False,
                )
                + "\n"
            )
    print(f"wrote {out} ({len(ds)} rows) — keep local; do not commit")
    return 0


if __name__ == "__main__":
    sys.exit(main())
