#!/usr/bin/env python3
import mysql.connector
from mysql.connector import Error

def test_database_query():
    """直接测试数据库查询"""
    
    try:
        # 连接到MySQL数据库
        connection = mysql.connector.connect(
            host='localhost',
            port=3306,
            database='prediction_system',
            user='root',
            password='root123456'
        )
        
        if connection.is_connected():
            cursor = connection.cursor()
            
            print("=== 测试数据库查询 ===")
            
            # 测试用户表字段
            print("\n1. 检查用户表结构:")
            cursor.execute("DESCRIBE users")
            columns = cursor.fetchall()
            for col in columns:
                print(f"  {col[0]} - {col[1]}")
            
            # 测试排行榜查询（使用正确的字段名）
            print("\n2. 测试排行榜查询:")
            try:
                cursor.execute("SELECT id, username, nickname, points FROM users ORDER BY points DESC, createdAt ASC LIMIT 5")
                users = cursor.fetchall()
                print("✅ 查询成功！")
                for i, user in enumerate(users, 1):
                    print(f"  {i}. {user[1]} ({user[2]}) - {user[3]}分")
            except Exception as e:
                print(f"❌ 查询失败: {e}")
                
            # 测试错误的字段名
            print("\n3. 测试错误的字段名:")
            try:
                cursor.execute("SELECT id, username FROM users ORDER BY points DESC, created_at ASC LIMIT 1")
                print("❌ 不应该成功")
            except Exception as e:
                print(f"✅ 预期的错误: {e}")
                
    except Error as e:
        print(f"MySQL error: {e}")
    except Exception as e:
        print(f"Error: {e}")
    finally:
        if connection and connection.is_connected():
            cursor.close()
            connection.close()

if __name__ == "__main__":
    test_database_query()