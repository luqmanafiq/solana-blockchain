"use client";

import { useEffect, useState } from "react";
import * as web3 from "@solana/web3.js";

// Phantom wallet typing
interface PhantomProvider {
  isPhantom?: boolean;
  publicKey?: web3.PublicKey;
  connect: () => Promise<{ publicKey: web3.PublicKey }>;
  disconnect: () => Promise<void>;
  signTransaction: (tx: web3.Transaction) => Promise<web3.Transaction>;
  signAllTransactions: (txs: web3.Transaction[]) => Promise<web3.Transaction[]>;
}

declare global {
  interface Window {
    solana?: PhantomProvider;
  }
}

export default function TransferPage() {
  const [wallet, setWallet] = useState<web3.PublicKey | null>(null);
  const [toAddress, setToAddress] = useState("");
  const [amount, setAmount] = useState("");
  const [solPrice, setSolPrice] = useState<number | null>(null);
  const [showUSD, setShowUSD] = useState(false);
  const [status, setStatus] = useState("");

  const connection = new web3.Connection(
    "https://devnet.helius-rpc.com/?api-key=0f803376-0189-4d72-95f6-a5f41cef157d"
  );

  // Fetch SOL price in USD
  useEffect(() => {
    fetch("https://api.coingecko.com/api/v3/simple/price?ids=solana&vs_currencies=usd")
      .then((res) => res.json())
      .then((data) => setSolPrice(data.solana.usd))
      .catch((err) => console.error("Error fetching price", err));
  }, []);

  // Connect Phantom wallet
  const connectWallet = async () => {
    if (window.solana?.isPhantom) {
      try {
        const resp = await window.solana.connect();
        setWallet(resp.publicKey);
        setStatus(`Connected: ${resp.publicKey.toBase58()}`);
      } catch (err) {
        console.error(err);
        setStatus("Wallet connection failed.");
      }
    } else {
      alert("Phantom wallet not found. Please install it.");
    }
  };

  // Validate address
  const isValidSolAddress = (address: string) => {
    try {
      new web3.PublicKey(address);
      return true;
    } catch {
      return false;
    }
  };

  // Send transaction
  const sendTransaction = async () => {
    if (!wallet) {
      setStatus("Connect your wallet first!");
      return;
    }
    if (!isValidSolAddress(toAddress)) {
      setStatus("Invalid recipient address.");
      return;
    }
    if (Number(amount) <= 0) {
      setStatus("Enter a valid amount.");
      return;
    }

    try {
      const lamports = web3.LAMPORTS_PER_SOL * parseFloat(amount);
      const transaction = new web3.Transaction().add(
        web3.SystemProgram.transfer({
          fromPubkey: wallet,
          toPubkey: new web3.PublicKey(toAddress),
          lamports,
        })
      );

      transaction.feePayer = wallet;
      transaction.recentBlockhash = (await connection.getLatestBlockhash()).blockhash;

      // Sign with Phantom
      const signedTx = await window.solana!.signTransaction(transaction);
      const txid = await connection.sendRawTransaction(signedTx.serialize());

      setStatus(`✅ Transaction sent! TxID: ${txid}`);
    } catch (err) {
      console.error(err);
      setStatus("❌ Transaction failed.");
    }
  };

  return (
    <div style={{ padding: "20px", fontFamily: "Arial, sans-serif" }}>
      <h1>💸 Transfer SOL</h1>

      {!wallet ? (
        <button onClick={connectWallet} style={{ padding: "10px", marginBottom: "20px" }}>
          Connect Phantom Wallet
        </button>
      ) : (
        <p>Connected: {wallet.toBase58()}</p>
      )}

      <div style={{ marginBottom: "10px" }}>
        <label>Amount (in SOL): </label>
        <input
          type="number"
          value={amount}
          onChange={(e) => setAmount(e.target.value)}
          style={{ padding: "5px" }}
        />
        {showUSD && solPrice && amount && (
          <span style={{ marginLeft: "10px" }}>
            ≈ ${(parseFloat(amount) * solPrice).toFixed(2)} USD
          </span>
        )}
      </div>

      <div style={{ marginBottom: "10px" }}>
        <label>To Wallet Address: </label>
        <input
          type="text"
          value={toAddress}
          onChange={(e) => setToAddress(e.target.value)}
          style={{ width: "400px", padding: "5px" }}
        />
      </div>

      <div style={{ marginBottom: "20px" }}>
        <label>
          <input
            type="checkbox"
            checked={showUSD}
            onChange={() => setShowUSD(!showUSD)}
          />{" "}
          Show USD Value
        </label>
      </div>

      <button onClick={sendTransaction} style={{ padding: "10px 20px" }}>
        Send SOL
      </button>

      {status && <p style={{ marginTop: "20px" }}>{status}</p>}
    </div>
  );
}
