import { encodeFunctionData, decodeFunctionResult } from "viem";

export class WalletManager {
    constructor() {
        this.ethereum = window.ethereum;
    }

    isInstalled() {
        return !!this.ethereum;
    }

    async getAccounts() {
        return await this.ethereum.request({ method: "eth_accounts" });
    }

    async getChainId() {
        return await this.ethereum.request({ method: "eth_chainId" });
    }

    async requestAccounts() {
        return await this.ethereum.request({ method: "eth_requestAccounts" });
    }

    async switchChain(chainId) {
        return await this.ethereum.request({
            method: "wallet_switchEthereumChain",
            params: [{ chainId }],
        });
    }

    async addChain(params) {
        return await this.ethereum.request({
            method: "wallet_addEthereumChain",
            params: [params],
        });
    }

    async revokePermissions() {
        return await this.ethereum.request({
            method: "wallet_revokePermissions",
            params: [{ eth_accounts: {} }],
        });
    }

    onChainChanged(callback) {
        this.ethereum.on("chainChanged", callback);
    }

    onAccountsChanged(callback) {
        this.ethereum.on("accountsChanged", callback);
    }

    removeAllListeners() {
        this.ethereum.removeAllListeners("chainChanged");
        this.ethereum.removeAllListeners("accountsChanged");
    }

    async callContract(to, data) {
        return await this.ethereum.request({
            method: "eth_call",
            params: [{ to, data }, "latest"],
        });
    }

    async sendTx(from, to, data) {
        return await this.ethereum.request({
            method: "eth_sendTransaction",
            params: [{ from, to, data }],
        });
    }

    async getBalance(address) {
        return await this.ethereum.request({
            method: "eth_getBalance",
            params: [address, "latest"],
        });
    }

    async getFeeHistory(blockCount, rewardPercentiles) {
        return await this.ethereum.request({
            method: "eth_feeHistory",
            params: [blockCount, "latest", rewardPercentiles],
        });
    }

    async personalSign(message, address) {
        return await this.ethereum.request({
            method: "personal_sign",
            params: [message, address],
        });
    }
}

export async function callContractFunction(wallet, contractAddress, abi, functionName, args = []) {
    const data = encodeFunctionData({ abi, functionName, args });
    const result = await wallet.callContract(contractAddress, data);
    return decodeFunctionResult({ abi, functionName, data: result });
}

export async function sendContractTx(wallet, from, contractAddress, abi, functionName, args = []) {
    const data = encodeFunctionData({ abi, functionName, args });
    return await wallet.sendTx(from, contractAddress, data);
}

function normalizeChainId(chainId) {
    return typeof chainId === "string" ? chainId.toLowerCase() : `0x${chainId.toString(16)}`;
}

export function createWalletConnection(callbacks) {
    const wallet = new WalletManager();

    async function initialize() {
        if (!wallet.isInstalled()) {
            callbacks.onError("MetaMask not detected. Please install MetaMask.");
            return;
        }

        wallet.onChainChanged((chainId) => {
            const isTestnet = normalizeChainId(chainId) === "0xaa36a7";
            callbacks.onChainChanged(isTestnet);
        });

        wallet.onAccountsChanged((accounts) => {
            if (accounts.length > 0) {
                callbacks.onAccountChanged(accounts[0]);
            } else {
                disconnect();
            }
        });

        try {
            const accounts = await wallet.getAccounts();
            if (accounts.length > 0) {
                const chainId = await wallet.getChainId();
                const isTestnet = normalizeChainId(chainId) === "0xaa36a7";
                callbacks.onConnected(accounts[0], isTestnet);
            }
        } catch (error) {
            callbacks.onError(`Error checking connection: ${error.message}`);
        }
    }

    async function connect(isTestnet) {
        try {
            await wallet.requestAccounts();
            const currentChainId = await wallet.getChainId();
            const desiredChainId = isTestnet ? "0xaa36a7" : "0x1";

            if (currentChainId.toLowerCase() !== desiredChainId.toLowerCase()) {
                try {
                    await wallet.switchChain(desiredChainId);
                } catch (switchError) {
                    if (switchError.code === 4902 && isTestnet) {
                        await wallet.addChain({
                            chainId: "0xaa36a7",
                            chainName: "Sepolia",
                            nativeCurrency: { name: "ETH", symbol: "ETH", decimals: 18 },
                            rpcUrls: ["https://rpc.sepolia.org"],
                            blockExplorerUrls: ["https://sepolia.etherscan.io"],
                        });
                    } else if (switchError.code === 4001) {
                        const actualChainId = await wallet.getChainId();
                        const actualIsTestnet = normalizeChainId(actualChainId) === "0xaa36a7";
                        const accounts = await wallet.getAccounts();
                        callbacks.onConnected(accounts[0], actualIsTestnet, true);
                        return;
                    }
                }
            }

            const finalChainId = await wallet.getChainId();
            const finalIsTestnet = normalizeChainId(finalChainId) === "0xaa36a7";
            const accounts = await wallet.getAccounts();
            callbacks.onConnected(accounts[0], finalIsTestnet);
        } catch (error) {
            const message =
                error.code === 4001 ? "Connection cancelled by user" : `Connection failed: ${error.message}`;
            callbacks.onError(message);
        }
    }

    async function disconnect() {
        try {
            await wallet.revokePermissions();
        } catch {}

        wallet.removeAllListeners();
        callbacks.onDisconnected();
    }

    return { wallet, initialize, connect, disconnect };
}