#!/usr/bin/env python3
"""
Google ADK Integration Example
Demonstrates how to create an AI agent with M-Pesa payment capabilities
"""

import os
import requests
from typing import Dict, Any


class MPesaTools:
    """M-Pesa tools for Google ADK agents"""
    
    def __init__(self, mcp_url: str = "http://localhost:8080"):
        self.mcp_url = mcp_url
        self.session_id = None
        self.message_endpoint = None
        self._connect()
    
    def _connect(self):
        """Connect to MCP server via SSE"""
        try:
            response = requests.get(f"{self.mcp_url}/sse", stream=True, timeout=10)
            for line in response.iter_lines():
                if line:
                    decoded = line.decode('utf-8')
                    if decoded.startswith('event: endpoint'):
                        continue
                    if decoded.startswith('data: '):
                        endpoint = decoded[6:]
                        self.message_endpoint = f"{self.mcp_url}{endpoint}"
                        self.session_id = endpoint.split('session=')[1]
                        print(f"Connected to MCP server: {self.session_id}")
                        break
        except Exception as e:
            print(f"Failed to connect to MCP server: {e}")
            raise
    
    def _call_mcp_tool(self, tool_name: str, arguments: Dict[str, Any]) -> Dict[str, Any]:
        """Call an MCP tool"""
        message = {
            "jsonrpc": "2.0",
            "id": 1,
            "method": "tools/call",
            "params": {
                "name": tool_name,
                "arguments": arguments
            }
        }
        
        response = requests.post(self.message_endpoint, json=message, timeout=30)
        return response.json()
    
    def charge_customer(self, phone_number: str, amount: int, description: str = "") -> str:
        """
        Charge a customer via M-Pesa STK Push
        
        Args:
            phone_number: Customer's M-Pesa number (254XXXXXXXXX or 0XXXXXXXXX)
            amount: Amount in KES
            description: Payment description (optional)
            
        Returns:
            Human-readable result message
        """
        try:
            result = self._call_mcp_tool("stk_push", {
                "phone_number": phone_number,
                "amount": amount
            })
            
            if "result" in result:
                content = result["result"].get("content", [{}])[0]
                return content.get("text", "Payment request sent successfully!")
            else:
                return f"Payment request failed: {result.get('error', 'Unknown error')}"
                
        except Exception as e:
            return f"Error processing payment: {str(e)}"
    
    def generate_payment_qr(
        self,
        merchant_name: str,
        amount: int,
        till_number: str,
        reference: str = "PAYMENT"
    ) -> str:
        """
        Generate a QR code for payment
        
        Args:
            merchant_name: Business name
            amount: Amount in KES
            till_number: Till number
            reference: Payment reference
            
        Returns:
            Human-readable result with QR code info
        """
        try:
            result = self._call_mcp_tool("generate_qr_code", {
                "merchant_name": merchant_name,
                "ref_no": reference,
                "amount": amount,
                "trx_code": "BG",  # Buy Goods
                "cp_identifier": till_number
            })
            
            if "result" in result:
                content = result["result"].get("content", [{}])[0]
                return content.get("text", "QR code generated successfully!")
            else:
                return f"QR generation failed: {result.get('error', 'Unknown error')}"
                
        except Exception as e:
            return f"Error generating QR code: {str(e)}"


def create_payment_agent():
    """
    Create a Google ADK agent with M-Pesa payment capabilities
    
    Note: This requires Google ADK to be installed:
    pip install google-adk
    """
    try:
        from google.adk import Agent
    except ImportError:
        print("Google ADK not installed. Install with: pip install google-adk")
        return None
    
    # Initialize M-Pesa tools
    mpesa = MPesaTools()
    
    # Create agent with M-Pesa capabilities
    agent = Agent(
        name="mpesa-payment-assistant",
        model="gemini-2.0-flash-exp",
        tools=[mpesa.charge_customer, mpesa.generate_payment_qr],
        instruction="""
        You are a helpful payment assistant that processes M-Pesa payments for customers.
        
        Your capabilities:
        1. Charge customers using STK Push (sends payment request to their phone)
        2. Generate QR codes for in-person payments
        
        When processing payments:
        - Always confirm the amount and phone number with the customer
        - Explain that they'll receive a prompt on their phone to enter M-Pesa PIN
        - Be patient and helpful if they have questions
        - For QR codes, explain that they should scan with their M-Pesa app
        
        Phone number format:
        - Accept: 254712345678 or 0712345678
        - The system will automatically format it correctly
        
        Example interactions:
        - "Charge 0712345678 KES 1000 for order #12345"
        - "Generate a QR code for KES 500 at My Store, till 123456"
        """
    )
    
    return agent


def main():
    """Example usage"""
    print("=" * 60)
    print("Google ADK + M-Pesa Integration Example")
    print("=" * 60)
    
    # Create agent
    print("\nCreating payment agent...")
    agent = create_payment_agent()
    
    if not agent:
        print("\nRunning without Google ADK (direct tool usage)...")
        mpesa = MPesaTools()
        
        # Example 1: Charge customer
        print("\n1. Charging customer...")
        result = mpesa.charge_customer(
            phone_number="254712345678",
            amount=1000,
            description="Order #12345"
        )
        print(f"Result: {result}")
        
        # Example 2: Generate QR code
        print("\n2. Generating QR code...")
        result = mpesa.generate_payment_qr(
            merchant_name="Test Store",
            amount=500,
            till_number="123456",
            reference="ORDER123"
        )
        print(f"Result: {result}")
        
    else:
        print("✓ Agent created successfully!")
        
        # Example conversations
        print("\n" + "=" * 60)
        print("Example Agent Conversations:")
        print("=" * 60)
        
        examples = [
            "Charge customer 254712345678 KES 1000 for order #12345",
            "Generate a QR code for KES 500 at My Store with till number 123456",
            "I need to collect KES 2500 from phone number 0722334455"
        ]
        
        for i, prompt in enumerate(examples, 1):
            print(f"\n{i}. User: {prompt}")
            try:
                response = agent.run(prompt)
                print(f"   Agent: {response}")
            except Exception as e:
                print(f"   Error: {e}")
    
    print("\n" + "=" * 60)
    print("Examples completed!")
    print("=" * 60)


if __name__ == "__main__":
    main()
