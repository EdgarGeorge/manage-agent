#!/usr/bin/env python3
"""
迁移脚本：从旧的数据格式迁移到新的数据格式

旧格式：
- worth_models 表有固定字段：cash, stock_a, stock_m, hongli, bond, debt

新格式：
- worth_models 表只有 time 字段
- type_worth_models 表存储各类型的值，使用 type_name 字段
- f_type_models 表存储资金类型定义
"""

import sqlite3
import os
from pathlib import Path

# 配置路径
script_dir = Path(__file__).parent
project_root = script_dir.parent
old_db_path = project_root / "manage-agent.db"
new_db_path = project_root / "bin" / "data" / "manage-agent.db"

# 旧格式的字段映射到新格式的类型名
OLD_FIELD_TO_TYPE_NAME = {
    "cash": "cash",
    "stock_a": "stock_a",
    "stock_m": "stock_m",
    "hongli": "hongli",
    "bond": "bond",
    "debt": "debt",
}

# 类型名到中文名的映射
TYPE_NAME_TO_CNAME = {
    "cash": "现金",
    "stock_a": "A股",
    "stock_m": "美股",
    "hongli": "红利品类",
    "bond": "债券",
    "debt": "债权",
}


def check_old_db_structure(old_conn):
    """检查旧数据库的表结构"""
    cursor = old_conn.cursor()
    cursor.execute("SELECT name FROM sqlite_master WHERE type='table'")
    tables = [row[0] for row in cursor.fetchall()]
    print(f"旧数据库中的表: {tables}")

    if "worth_models" in tables:
        cursor.execute("PRAGMA table_info(worth_models)")
        columns = cursor.fetchall()
        print(f"\nworth_models 表的列:")
        for col in columns:
            print(f"  - {col[1]} ({col[2]})")

    return tables


def migrate_f_types(new_conn):
    """迁移资金类型表（确保所有类型都存在）"""
    cursor = new_conn.cursor()

    for type_name, cname in TYPE_NAME_TO_CNAME.items():
        # 检查是否已存在
        cursor.execute("SELECT id FROM f_type_models WHERE name = ?", (type_name,))
        if cursor.fetchone():
            print(f"  类型 {type_name} 已存在，跳过")
            continue

        # 插入新类型
        cursor.execute(
            "INSERT INTO f_type_models (created_at, updated_at, deleted_at, name, cname, `order`) VALUES (datetime('now'), datetime('now'), NULL, ?, ?, ?)",
            (type_name, cname, 0),
        )
        print(f"  创建类型: {type_name} ({cname})")

    new_conn.commit()


def migrate_worth_data(old_conn, new_conn):
    """迁移现值数据"""
    old_cursor = old_conn.cursor()
    new_cursor = new_conn.cursor()

    # 检查旧表是否有 sync_status 字段
    old_cursor.execute("PRAGMA table_info(worth_models)")
    columns = [col[1] for col in old_cursor.fetchall()]
    has_sync_status = "sync_status" in columns

    # 构建查询语句（根据字段是否存在）
    if has_sync_status:
        select_sql = """
            SELECT id, created_at, updated_at, deleted_at, time, cash, stock_a, stock_m, hongli, bond, debt, sync_status
            FROM worth_models
            ORDER BY id
        """
    else:
        select_sql = """
            SELECT id, created_at, updated_at, deleted_at, time, cash, stock_a, stock_m, hongli, bond, debt
            FROM worth_models
            ORDER BY id
        """

    # 从旧数据库读取数据
    old_cursor.execute(select_sql)

    old_rows = old_cursor.fetchall()
    print(f"\n找到 {len(old_rows)} 条旧的现值记录")

    if not old_rows:
        print("没有数据需要迁移")
        return

    migrated_count = 0

    for old_row in old_rows:
        if has_sync_status:
            (
                old_id,
                created_at,
                updated_at,
                deleted_at,
                time,
                cash,
                stock_a,
                stock_m,
                hongli,
                bond,
                debt,
                sync_status,
            ) = old_row
        else:
            (
                old_id,
                created_at,
                updated_at,
                deleted_at,
                time,
                cash,
                stock_a,
                stock_m,
                hongli,
                bond,
                debt,
            ) = old_row
            sync_status = 0  # 默认值

        # 检查新数据库中是否已存在（根据 time）
        new_cursor.execute("SELECT id FROM worth_models WHERE time = ?", (time,))
        existing = new_cursor.fetchone()

        if existing:
            print(f"  跳过已存在的记录: time={time}")
            continue

        # 插入新的 worth_models 记录
        new_cursor.execute(
            """
            INSERT INTO worth_models (created_at, updated_at, deleted_at, time, sync_status)
            VALUES (?, ?, ?, ?, ?)
        """,
            (created_at, updated_at, deleted_at, time, sync_status or 0),
        )

        new_worth_id = new_cursor.lastrowid

        # 插入关联的 type_worth_models 记录
        type_values = {
            "cash": cash or 0.0,
            "stock_a": stock_a or 0.0,
            "stock_m": stock_m or 0.0,
            "hongli": hongli or 0.0,
            "bond": bond or 0.0,
            "debt": debt or 0.0,
        }

        for type_name, value in type_values.items():
            if value != 0.0:  # 只插入非零值
                new_cursor.execute(
                    """
                    INSERT INTO type_worth_models (created_at, updated_at, deleted_at, worth_id, type_name, value)
                    VALUES (datetime('now'), datetime('now'), NULL, ?, ?, ?)
                """,
                    (new_worth_id, type_name, value),
                )

        migrated_count += 1
        if migrated_count % 10 == 0:
            print(f"  已迁移 {migrated_count} 条记录...")

    new_conn.commit()
    print(f"\n✅ 成功迁移 {migrated_count} 条现值记录")


def migrate_flow_records(old_conn, new_conn):
    """迁移流水记录（格式不变）"""
    old_cursor = old_conn.cursor()
    new_cursor = new_conn.cursor()

    # 检查旧表是否存在
    old_cursor.execute(
        "SELECT name FROM sqlite_master WHERE type='table' AND name='flow_record_models'"
    )
    if not old_cursor.fetchone():
        print("\n旧数据库中没有 flow_record_models 表，跳过")
        return

    # 从旧数据库读取数据
    old_cursor.execute(
        """
        SELECT id, created_at, updated_at, deleted_at, time, type, value, sync_status
        FROM flow_record_models
        ORDER BY id
    """
    )

    old_rows = old_cursor.fetchall()
    print(f"\n找到 {len(old_rows)} 条流水记录")

    if not old_rows:
        print("没有流水记录需要迁移")
        return

    migrated_count = 0

    for old_row in old_rows:
        (
            old_id,
            created_at,
            updated_at,
            deleted_at,
            time,
            type_name,
            value,
            sync_status,
        ) = old_row

        # 检查新数据库中是否已存在
        new_cursor.execute(
            """
            SELECT id FROM flow_record_models 
            WHERE time = ? AND type = ? AND value = ?
        """,
            (time, type_name, value),
        )

        if new_cursor.fetchone():
            continue

        # 插入新记录
        new_cursor.execute(
            """
            INSERT INTO flow_record_models (created_at, updated_at, deleted_at, time, type, value, sync_status)
            VALUES (?, ?, ?, ?, ?, ?, ?)
        """,
            (
                created_at,
                updated_at,
                deleted_at,
                time,
                type_name,
                value,
                sync_status or 0,
            ),
        )

        migrated_count += 1

    new_conn.commit()
    print(f"✅ 成功迁移 {migrated_count} 条流水记录")


def main():
    print("=" * 60)
    print("数据迁移脚本：从旧格式迁移到新格式")
    print("=" * 60)

    # 检查文件是否存在
    if not old_db_path.exists():
        print(f"❌ 错误: 旧数据库文件不存在: {old_db_path}")
        return

    if not new_db_path.exists():
        print(f"❌ 错误: 新数据库文件不存在: {new_db_path}")
        print("请先运行一次应用程序以创建新数据库")
        return

    print(f"\n旧数据库: {old_db_path}")
    print(f"新数据库: {new_db_path}")

    try:
        # 连接数据库
        print("\n正在连接数据库...")
        old_conn = sqlite3.connect(str(old_db_path))
        new_conn = sqlite3.connect(str(new_db_path))

        # 检查旧数据库结构
        print("\n检查旧数据库结构...")
        old_tables = check_old_db_structure(old_conn)

        if "worth_models" not in old_tables:
            print("错误: 旧数据库中没有 worth_models 表")
            return

        # 开始迁移
        print("\n" + "=" * 60)
        print("开始迁移数据...")
        print("=" * 60)

        # 1. 迁移资金类型
        print("\n1. 迁移资金类型...")
        migrate_f_types(new_conn)

        # 2. 迁移现值数据
        print("\n2. 迁移现值数据...")
        migrate_worth_data(old_conn, new_conn)

        # 3. 迁移流水记录
        print("\n3. 迁移流水记录...")
        migrate_flow_records(old_conn, new_conn)

        print("\n" + "=" * 60)
        print("数据迁移完成！")
        print("=" * 60)

    except Exception as e:
        print(f"\n迁移过程中出错: {e}")
        import traceback

        traceback.print_exc()
    finally:
        if "old_conn" in locals():
            old_conn.close()
        if "new_conn" in locals():
            new_conn.close()
        print("\n数据库连接已关闭")


if __name__ == "__main__":
    main()
