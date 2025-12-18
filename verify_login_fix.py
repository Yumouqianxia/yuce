#!/usr/bin/env python3
import requests
import json

def test_login():
    """测试登录功能是否修复"""
    
    # 测试用户（请根据实际情况调整密码）
    test_users = [
        {'username': 'root', 'password': 'root123456'},  # 管理员账户
        {'username': 'mengshang004', 'password': '123456'},  # 普通用户
    ]
    
    base_url = "http://localhost:1874/api/v1"
    
    print("=== 登录功能测试 ===")
    
    for user in test_users:
        print(f"\n测试用户: {user['username']}")
        
        try:
            # 发送登录请求
            response = requests.post(
                f"{base_url}/auth/login",
                json={
                    "username": user['username'],
                    "password": user['password']
                },
                headers={'Content-Type': 'application/json'},
                timeout=10
            )
            
            print(f"HTTP状态码: {response.status_code}")
            
            if response.status_code == 200:
                data = response.json()
                print("✅ 登录成功！")
                print(f"用户ID: {data.get('user', {}).get('id')}")
                print(f"用户名: {data.get('user', {}).get('username')}")
                print(f"昵称: {data.get('user', {}).get('nickname')}")
                print(f"角色: {data.get('user', {}).get('role')}")
                print(f"积分: {data.get('user', {}).get('points')}")
                print(f"访问令牌: {data.get('access_token', '')[:50]}...")
            else:
                print("❌ 登录失败")
                try:
                    error_data = response.json()
                    print(f"错误信息: {error_data.get('message', '未知错误')}")
                except:
                    print(f"响应内容: {response.text}")
                    
        except requests.exceptions.ConnectionError:
            print("❌ 无法连接到后端服务，请确保后端正在运行")
        except requests.exceptions.Timeout:
            print("❌ 请求超时")
        except Exception as e:
            print(f"❌ 请求出错: {e}")

def check_backend_status():
    """检查后端服务状态"""
    try:
        response = requests.get("http://localhost:1874/health", timeout=5)
        if response.status_code == 200:
            print("✅ 后端服务正在运行")
            return True
        else:
            print(f"⚠️ 后端服务状态异常: {response.status_code}")
            return False
    except:
        print("❌ 后端服务未运行，请先启动后端服务")
        return False

if __name__ == "__main__":
    print("=== 登录修复验证脚本 ===")
    
    # 检查后端服务状态
    if check_backend_status():
        test_login()
    else:
        print("\n请先启动后端服务：")
        print("cd backend-go && go run cmd/api/main.go")
        print("或者使用Docker：")
        print("docker-compose up backend")