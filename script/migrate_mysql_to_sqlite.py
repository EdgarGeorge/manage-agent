import mysql.connector
import sqlite3
import os
from datetime import datetime
from decimal import Decimal

# --- 配置 ---
# MySQL 连接配置
mysql_config = {
    "host": "127.0.0.1",
    "user": "root",
    "password": "root",
    "database": "manage-agent",
}

# SQLite 数据库路径
script_dir = os.path.dirname(os.path.abspath(__file__))
sqlite_path = os.path.join(script_dir, "..", "bin", "data", "manage-agent.db")

# 需要迁移的表名
tables_to_migrate = ["request_record_models", "worth_models", "flow_record_models"]


def convert_value(value):
    """转换值为 SQLite 兼容的格式"""
    if value is None:
        return None
    # 处理 Decimal 类型
    if isinstance(value, Decimal):
        return float(value)
    # 处理 datetime 对象
    if isinstance(value, datetime):
        return value.strftime("%Y-%m-%d %H:%M:%S")
    # 处理其他可能不兼容的类型
    return value


def migrate():
    # 检查 SQLite 文件是否存在
    if not os.path.exists(sqlite_path):
        print(f"错误: SQLite 数据库文件不存在于 '{os.path.abspath(sqlite_path)}'")
        print(
            "请先运行一次 Go 应用程序 (go run bin/main.go) 来自动创建数据库和表结构。"
        )
        return

    try:
        # 连接数据库
        print("正在连接 MySQL...")
        mysql_conn = mysql.connector.connect(**mysql_config)
        mysql_cursor = mysql_conn.cursor(dictionary=True)

        print("正在连接 SQLite...")
        sqlite_conn = sqlite3.connect(sqlite_path)
        sqlite_conn.row_factory = sqlite3.Row
        sqlite_cursor = sqlite_conn.cursor()

        # 迁移数据
        for table in tables_to_migrate:
            print(f"\n--- 开始迁移表: {table} ---")

            # 1. 从 MySQL 读取数据
            mysql_cursor.execute(f"SELECT * FROM `{table}`")
            rows = mysql_cursor.fetchall()

            if not rows:
                print(f"  表 '{table}' 在 MySQL 中没有数据，跳过。")
                continue

            print(f"  在 MySQL 中找到 {len(rows)} 条记录。")

            # 2. 清空 SQLite 中的目标表
            try:
                sqlite_cursor.execute(f"DELETE FROM `{table}`")
                print(f"  已清空 SQLite 中的表 '{table}'。")
            except sqlite3.OperationalError:
                print(f"  警告: SQLite 中不存在表 '{table}'，将直接插入。")
                continue

            # 3. 获取列信息
            columns = list(rows[0].keys())
            columns_str = ", ".join([f"`{col}`" for col in columns])
            placeholders = ", ".join(["?"] * len(columns))

            # 4. 准备数据并转换类型
            data_to_insert = []
            for row in rows:
                try:
                    # 转换每列的值
                    values = [convert_value(row[col]) for col in columns]
                    data_to_insert.append(values)
                except Exception as e:
                    print(f"  警告: 转换数据时出错: {e}")
                    print(f"  表: {table}, 行: {row}")
                    continue

            if not data_to_insert:
                print("  没有有效数据可以插入，跳过。")
                continue

            # 5. 批量插入数据
            try:
                sqlite_cursor.executemany(
                    f"INSERT INTO `{table}` ({columns_str}) VALUES ({placeholders})",
                    data_to_insert,
                )
                sqlite_conn.commit()
                print(f"  成功插入 {len(data_to_insert)} 条记录。")
            except sqlite3.Error as e:
                sqlite_conn.rollback()
                print(f"  错误: 插入数据时出错: {e}")
                print(f"  表: {table}")
                print(
                    f"  SQL: INSERT INTO {table} ({columns_str}) VALUES ({placeholders})"
                )
                if data_to_insert:
                    print(f"  示例数据: {data_to_insert[0]}")

        print("\n🎉 数据迁移完成！")

    except mysql.connector.Error as err:
        print(f"MySQL 错误: {err}")
    except sqlite3.Error as err:
        print(f"SQLite 错误: {err}")
    except Exception as e:
        print(f"未知错误: {e}")
    finally:
        # 关闭连接
        if "mysql_conn" in locals() and mysql_conn.is_connected():
            mysql_conn.close()
            print("MySQL 连接已关闭。")
        if "sqlite_conn" in locals():
            sqlite_conn.close()
            print("SQLite 连接已关闭。")


if __name__ == "__main__":
    migrate()
