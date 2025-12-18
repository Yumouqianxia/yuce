#!/usr/bin/env python3
import subprocess
import json

def test_leaderboard_api():
    """测试排行榜API"""
    
    # 测试不同的排行榜端点
    endpoints = [
        {
            "name": "全局排行榜",
            "url": "http://localhost:1874/api/leaderboard",
            "params": ""
        },
        {
            "name": "夏季锦标赛排行榜",
            "url": "http://localhost:1874/api/leaderboard",
            "params": "?tournament=SUMMER&limit=5"
        },
        {
            "name": "排行榜统计",
            "url": "http://localhost:1874/api/leaderboard/stats",
            "params": ""
        },
        {
            "name": "用户排名",
            "url": "http://localhost:1874/api/leaderboard/users/1/rank",
            "params": ""
        }
    ]
    
    print("=== 排行榜API测试 ===")
    
    for endpoint in endpoints:
        print(f"\n🔍 测试: {endpoint['name']}")
        test_endpoint(endpoint['url'] + endpoint['params'])

def test_endpoint(url):
    """测试单个端点"""
    
    ps_cmd = f'''
    try {{
        $response = Invoke-WebRequest -Uri "{url}" -Method GET
        Write-Host "✅ 状态码: $($response.StatusCode)"
        $data = $response.Content | ConvertFrom-Json
        if ($data.success) {{
            Write-Host "✅ 请求成功: $($data.message)"
            if ($data.data -is [array]) {{
                Write-Host "  返回数据条数: $($data.data.Count)"
                if ($data.data.Count -gt 0) {{
                    $first = $data.data[0]
                    if ($first.username) {{
                        Write-Host "  第一名: $($first.username) ($($first.nickname)) - $($first.points)分"
                    }}
                }}
            }} else {{
                Write-Host "  返回数据类型: $($data.data.GetType().Name)"
            }}
        }} else {{
            Write-Host "❌ 请求失败: $($data.message)"
        }}
    }} catch {{
        Write-Host "❌ 请求出错"
        Write-Host "  错误: $($_.Exception.Message)"
        if ($_.Exception.Response) {{
            $statusCode = $_.Exception.Response.StatusCode.value__
            Write-Host "  状态码: $statusCode"
            if ($statusCode -eq 500) {{
                Write-Host "  这是服务器内部错误，请检查后端日志"
            }}
        }}
    }}
    '''
    
    try:
        result = subprocess.run(
            ["powershell", "-Command", ps_cmd],
            capture_output=True,
            text=True,
            timeout=10
        )
        
        print(result.stdout)
        if result.stderr:
            print(f"错误: {result.stderr}")
            
    except Exception as e:
        print(f"测试出错: {e}")

if __name__ == "__main__":
    test_leaderboard_api()