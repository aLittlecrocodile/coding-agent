"""
循环驱动脚本示例：串联 prompts/00-04 与 Claude，形成自维护循环。

使用前：
- 设置环境变量 ANTHROPIC_API_KEY（必需）。
- 可选设置 LOOP_GOAL、CLAUDE_MODEL（默认 claude-3-opus-20240229）、MAX_ITER（默认 10）。
- 本脚本默认不执行任何命令，只记录 executor 输出。需要自动执行时请自行接入并做好二次确认。
"""

from __future__ import annotations

import json
import os
import sys
import time
from dataclasses import dataclass
from pathlib import Path
from typing import Any, Dict, List, Optional

try:
    from anthropic import Anthropic
except Exception as e:  # pragma: no cover
    sys.stderr.write("请先安装 anthropic：pip install anthropic\n")
    raise

BASE_DIR = Path(__file__).parent
PROMPTS_DIR = BASE_DIR / "prompts"
RUNS_DIR = BASE_DIR / "runs"
RUNS_DIR.mkdir(exist_ok=True)


def read_prompt(name: str) -> str:
    return (PROMPTS_DIR / name).read_text(encoding="utf-8")


def extract_json(text: str) -> Dict[str, Any]:
    """
    从模型回复中提取 JSON（容忍前后多余文本或代码块）。
    """
    first = text.find("{")
    last = text.rfind("}")
    if first == -1 or last == -1 or last < first:
        raise ValueError("未找到 JSON 对象")
    snippet = text[first : last + 1]
    return json.loads(snippet)


def ensure_schema(obj: Dict[str, Any], required: List[str]) -> Dict[str, Any]:
    if not isinstance(obj, dict):
        raise ValueError("输出不是 JSON 对象")
    for key in required:
        if key not in obj:
            obj[key] = [] if key.endswith("s") else ""
    return obj


@dataclass
class StepResult:
    name: str
    raw: str
    data: Dict[str, Any]


class LoopDriver:
    def __init__(self) -> None:
        api_key = os.getenv("ANTHROPIC_API_KEY")
        if not api_key:
            raise RuntimeError("缺少环境变量 ANTHROPIC_API_KEY")
        self.client = Anthropic(api_key=api_key)
        self.model = os.getenv("CLAUDE_MODEL", "claude-3-opus-20240229")
        self.max_iter = int(os.getenv("MAX_ITER", "10"))
        goal = os.getenv("LOOP_GOAL", "").strip()
        if not goal:
            goal = "请在环境变量 LOOP_GOAL 中填写目标"
        self.state: Dict[str, Any] = {
            "iteration": 0,
            "goal": goal,
            "plan": [],
            "actions": [],
            "notes": [],
            "results": "",
            "stop_reason": None,
        }
        ts = time.strftime("%Y%m%d-%H%M%S")
        self.log_path = RUNS_DIR / f"run-{ts}.jsonl"

    def log(self, event: str, payload: Dict[str, Any]) -> None:
        entry = {"event": event, "payload": payload}
        self.log_path.write_text(
            (self.log_path.read_text(encoding="utf-8") if self.log_path.exists() else "")
            + json.dumps(entry, ensure_ascii=False)
            + "\n",
            encoding="utf-8",
        )

    def call_step(self, step: str, extra: Optional[str] = None, required: Optional[List[str]] = None) -> StepResult:
        system = read_prompt("00_system_constraints.md")
        prompt_body = read_prompt(step)
        user_parts = [
            prompt_body,
            "STATE:",
            json.dumps(self.state, ensure_ascii=False),
        ]
        if extra:
            user_parts.append(extra)
        user_content = "\n\n".join(user_parts)

        # 最多尝试两次纠正格式
        attempts = 0
        last_raw = ""
        while attempts < 2:
            attempts += 1
            resp = self.client.messages.create(
                model=self.model,
                max_tokens=4000,
                system=system,
                messages=[{"role": "user", "content": user_content}],
            )
            raw = resp.content[0].text
            last_raw = raw
            try:
                data = extract_json(raw)
                if required:
                    data = ensure_schema(data, required)
                return StepResult(step, raw, data)
            except Exception:
                # 生成纠正提示
                user_content = (
                    f"{prompt_body}\n\nSTATE:\n{json.dumps(self.state, ensure_ascii=False)}\n\n"
                    "上一次回复未能解析为 JSON。请只返回 JSON 对象，遵守 schema。"
                )

        raise RuntimeError(f"{step} 连续输出非 JSON，最后输出: {last_raw[:200]}")

    def run(self) -> None:
        for i in range(self.max_iter):
            self.state["iteration"] = i

            planner = self.call_step(
                "01_planner.prompt.md",
                required=["plan", "next_actions", "risks", "assumptions"],
            )
            self.log("planner", planner.data)

            executor = self.call_step(
                "02_executor.prompt.md",
                extra=f"PLAN:\n{json.dumps(planner.data, ensure_ascii=False)}",
                required=["commands", "expected_outcomes", "fallbacks"],
            )
            self.log("executor", executor.data)

            # TODO: 在此处根据 executor.data['commands'] 进行人工确认与实际执行，收集输出。
            # 当前仅记录命令，不执行。

            reviewer = self.call_step(
                "03_reviewer.prompt.md",
                extra=(
                    f"PLAN:\n{json.dumps(planner.data, ensure_ascii=False)}\n\n"
                    f"EXEC:\n{json.dumps(executor.data, ensure_ascii=False)}"
                ),
                required=["findings", "deltas", "suggested_changes"],
            )
            self.log("reviewer", reviewer.data)

            controller = self.call_step(
                "04_loop_controller.prompt.md",
                extra=f"REVIEW:\n{json.dumps(reviewer.data, ensure_ascii=False)}",
                required=["decision", "reason", "plan_patch"],
            )
            self.log("controller", controller.data)

            # 更新 state 摘要
            self.state.update(
                {
                    "plan": planner.data.get("plan", []),
                    "actions": executor.data.get("commands", []),
                    "notes": reviewer.data.get("findings", []),
                    "results": reviewer.data.get("deltas", []),
                    "stop_reason": controller.data.get("reason"),
                }
            )

            decision = controller.data.get("decision", "stop")
            if decision in ("stop", "pause"):
                print(f"停止循环，原因: {controller.data.get('reason')}")
                break
            if decision == "adjust_plan" and controller.data.get("plan_patch"):
                # 简单地把 plan_patch 拼到 plan，实际可做更复杂的合并
                self.state["plan"] = controller.data.get("plan_patch", []) + self.state.get("plan", [])

        else:
            print("达到最大迭代次数后退出。")


def main() -> None:
    driver = LoopDriver()
    driver.run()
    print(f"运行日志保存在: {driver.log_path}")


if __name__ == "__main__":
    main()
