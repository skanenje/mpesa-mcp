#!/usr/bin/env python3
"""
M-Pesa MCP Client - Python Example
Demonstrates how to integrate with the M-Pesa MCP server from Python
"""

import requests
import json
import time
from typing import Optional, Dict, Any
import sseclient  # pip install sseclient-py


class MPesaMCPClient:
    """Client for interacting with M-Pesa MCP Server via SSE transport"""
    
    def __init__(self, base_url: str = "http://localhost:8080"):
        self.base_url = base_url
        self.session_id: Optional[str] = None
        self.message_endpoint: Optional[str] = None
        self.request_id = 0
        
    def connect(self) -> bool:
        """Establish SSE connection and get session ID"""
        try:
            print(f"Connecting to {self.base_url}/sse...")
            response = requests.get(f"{self.base_url}/sse", stream=True, timeout=10)
            client = sseclient.SSEClient(response)
            
            for event in client.events():
                if event.event == "endpoint":
                    self.message_endpoint = f"{self.base_url}{event.data}"
                    self.session_id = event.data.split("session=")[1]
                    print(f"✓ Connected! Session: {self.session_id}")
                    return True
                    
        except Exception as e:
            print(f"✗ Connection failed: {e}")
            return False
            
    def _call_tool(self, tool_name: str, arguments: Dict[str, Any]) -> Dict[str, Any]:
        """Internal method to call an MCP tool"""
        if not self.message_endpoint:
            raise Exception("Not connected. Call connect() first.")
            
        self.request_id += 1
        
        message = {
            "jsonrpc": "2.0",
            "id": self.request_id,
            "method": "tools/call",
            "params": {
                "name": tool_name,
                "arguments": arguments
            }
        }
        
        print(f"\n→ Calling tool: {tool_name}")
        print(f"  Arguments: {json.dumps(arguments, indent=2)}")
        
        response = requests.post(
            self.message_endpoint,
            json=message,
            headers={"Content-Type": "application/json"},
            timeout=30
        )
        
        if response.status_code == 202:
            print("  Status: Request accepted, processing...")
            # In production, you'd listen to SSE for the response
            # For this example, we'll just return the acceptance
            return {"status": "accepted", "request_id": self.request_id}
        else:
            print(f"  Status: {response.status_code}")
            return response.json()
    
    def stk_push(self, phone_number: str, amount: int) -> Dict[str, Any]:
        """
        Initiate STK Push payment
        
        Args:
            phone_number: Customer phone (254XXXXXXXXX or 0XXXXXXXXX)
            amount: Amount in KES
            
        Returns:
            Response from M-Pesa API
        """
        return self._call_tool("stk_push", {
            "phone_number": phone_number,
            "amount": amount
        })
    
    def generate_qr_code(
        self,
        merchant_name: str,
        amount: int,
        trx_code: str,
        cp_identifier: str,
        ref_no: str = "QR_PAYMENT"
    ) -> Dict[str, Any]:
        """
        Generate M-Pesa QR code
        
        Args:
            merchant_name: Business name
            amount: Amount in KES
            trx_code: Transaction type (BG, PB, SM, SB, WA)
            cp_identifier: Till number, paybill, or phone number
            ref_no: Transaction reference
            
        Returns:
            Response with QR code (base64 encoded)
        """
        return self._call_tool("generate_qr_code", {
            "merchant_name": merchant_name,
            "ref_no": ref_no,
            "amount": amount,
            "trx_code": trx_code,
            "cp_identifier": cp_identifier
        })
    
    def get_token_status(self) -> Dict[str, Any]:
        """Check OAuth token status"""
        return self._call_tool("get_token_status", {})
    
    def health_check(self) -> bool:
        """Check if server is healthy"""
        try:
            response = requests.get(f"{self.base_url}/health", timeout=5)
            return response.status_code == 200
        except:
            return False


def main():
    """Example usage"""
    print("=" * 60)
    print("M-Pesa MCP Client - Python Example")
    print("=" * 60)
    
    # Initialize client
    client = MPesaMCPClient()
    
    # Check server health
    print("\n1. Checking server health...")
    if not client.health_check():
        print("✗ Server is not running. Start it with: ./mpesa-mcp")
        return
    print("✓ Server is healthy")
    
    # Connect to SSE
    print("\n2. Connecting to MCP server...")
    if not client.connect():
        print("✗ Failed to connect")
        return
    
    # Example 1: Check token status
    print("\n3. Checking OAuth token status...")
    result = client.get_token_status()
    print(f"  Result: {json.dumps(result, indent=2)}")
    
    # Example 2: STK Push
    print("\n4. Initiating STK Push payment...")
    print("  (Using test number - no actual charge will occur in sandbox)")
    result = client.stk_push(
        phone_number="254723975141",
        amount=100
    )
    print(f"  Result: {json.dumps(result, indent=2)}")
    
    # Example 3: Generate QR Code
    print("\n5. Generating QR code...")
    result = client.generate_qr_code(
        merchant_name="Test Store",
        amount=500,
        trx_code="BG",  # Buy Goods
        cp_identifier="123456",  # Till number
        ref_no="ORDER123"
    )
    print(f"  Result: {json.dumps(result, indent=2)}")
    
    print("\n" + "=" * 60)
    print("Examples completed!")
    print("=" * 60)


if __name__ == "__main__":
    main()
