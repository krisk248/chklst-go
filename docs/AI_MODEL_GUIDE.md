# Choosing the Right AI Model for chklst

chklst uses a **local** Ollama model for its AI features (daily summary, analytics
narratives). This guide explains how to pick the right one for **your machine and task**
so you get good output without it being slow.

## The one rule that matters on a CPU machine

> **Speed depends on _active_ parameters, not total size. Quality depends on the model family + total knowledge.**

This laptop has **no usable GPU** (Intel Iris Xe, integrated) so everything runs on **CPU**.
On CPU, every token costs compute proportional to the **active** parameters per token:

- A **dense** model (e.g. `gemma4:12b`) activates **all 12B** params for every token → slow.
- A **Mixture-of-Experts (MoE)** model (e.g. `gemma4:26b`) activates only **~3.8B** params
  per token (the rest sit in RAM unused) → **fast like a 4B model, smart like a 26B model.**

Two separate costs to keep in mind:
| Cost | Driven by | On this laptop |
|------|-----------|----------------|
| **Cold load** (first call) | total file size on disk | ~1 min per GB; 18 GB ≈ slow first call |
| **Warm generation** | active params | dense 12B ≈ 3.3 tok/s; MoE 26B (4B active) ≈ 7–9 tok/s |
| **RAM resident** | total file size | must fit in free RAM (you have ~32 GB free) |

So: keep the model **warm** (pinned in RAM) and prefer **MoE** for the best speed/quality trade.

## Gemma 4 family at a glance (text tasks)

| Model | Size | Active params | CPU speed | Quality | Verdict for chklst |
|-------|------|---------------|-----------|---------|--------------------|
| `gemma4:e2b` | 7.2 GB | ~2B | fastest | basic | Too weak for good summaries |
| `gemma4:e4b` | 9.6 GB | ~4B | fast | good | **Lightweight pick** — snappy, solid |
| `gemma4:12b` (dense) | 7.6 GB | 12B | slow (~3 tok/s) | very good | Quality good but slow on CPU |
| **`gemma4:26b` (MoE)** | 18 GB | **3.8B** | **fast (~7–9 tok/s)** | **near-frontier** | ✅ **Recommended** — best quality at CPU-friendly speed |
| `gemma4:31b` (dense) | 20 GB | 31B | very slow | best | Too slow on CPU |
| `*-cloud` | — | — | network | best | Not local (data leaves machine) — avoid |

## Recommendation for chklst on this laptop

**Primary: `gemma4:26b`** — the MoE design gives you ~31B-class quality while only doing
~4B of compute per token, so it generates *faster* than the dense 12B you tested, with
better writing. It needs 18 GB RAM resident (you have it). The only cost is a slower
**cold** first call, which we hide by keeping the model warm / pre-warming before the 6 PM
daily-summary run.

**If you want it lighter/snappier: `gemma4:e4b`** — 4B effective, loads fast, still writes
good summaries. Good if you'd rather not give up 18 GB of RAM.

**Avoid on CPU:** dense `gemma4:12b` / `gemma4:31b` (all params active = slow), and the
`*-cloud` variants (they send data off-machine, defeating the local-AI goal).

## Task → model cheat-sheet

| Task in chklst | Needs | Best model |
|----------------|-------|-----------|
| Daily activity email (short, structured) | clean writing, speed | `gemma4:26b` (or `e4b`) |
| Analytics narrative / bad-patch reasoning | reasoning + summarizing | `gemma4:26b` |
| Quick/cheap experiments | speed over quality | `gemma4:e4b` |

## Settings that matter

- **Thinking mode OFF for summaries.** Gemma 4 can "think" (emit reasoning tokens first),
  which is slower. For structured summaries chklst keeps thinking disabled for speed and
  predictable output.
- **Keep-alive.** To avoid the multi-minute cold load, run Ollama so the model stays in RAM:
  `OLLAMA_KEEP_ALIVE=-1 ollama serve` (pins it), or chklst pre-warms before scheduled runs.
- **Sampling.** Google suggests `temperature=1.0, top_p=0.95, top_k=64`; chklst lowers the
  temperature for the daily email so the format stays consistent. All configurable in Settings.

## How to switch models

1. Pull it once: `ollama pull gemma4:26b`
2. In chklst → **Settings → AI (Ollama)**, set **Model** to `gemma4:26b`, click **Test AI**.

That's it — no rebuild. (`AI.sh` also takes `MODEL=gemma4:26b ./AI.sh`.)
