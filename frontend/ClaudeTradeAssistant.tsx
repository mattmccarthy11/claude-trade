import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { authApi } from '../lib/auth';
import { TurnstileWidget } from './TurnstileWidget';

interface TradeRecommendation {
  ticker: string;
  strategy: string;
  legs: string;
  thesis: string;
  pop: number;
  maxLoss: number;
  maxProfit: number;
  score: number;
}

interface ClaudeConnection {
  id: string;
  isActive: boolean;
  createdAt: Date;
  lastUsedAt?: Date;
}

export function ClaudeTradeAssistant() {
  const [showConnect, setShowConnect] = useState(false);
  const [apiKey, setApiKey] = useState('');
  const [showTurnstile, setShowTurnstile] = useState(false);
  const [turnstileToken, setTurnstileToken] = useState('');
  const queryClient = useQueryClient();
  const turnstileSiteKey = import.meta.env.VITE_TURNSTILE_SITE_KEY || '1x00000000000000000000AA';

  // Query to check Claude connection status
  const { data: connectionData, isLoading: isLoadingConnection } = useQuery({
    queryKey: ['claude-connection'],
    queryFn: async () => {
      const response = await authApi.get('/api/claude-code/status');
      if (!response.ok) throw new Error('Failed to fetch connection status');
      return response.json();
    },
  });

  // Query to get trade recommendations
  const { data: recommendations, isLoading: isLoadingRecs, refetch: refetchRecommendations } = useQuery({
    queryKey: ['trade-recommendations'],
    queryFn: async () => {
      const response = await authApi.get('/api/claude-code/recommendations');
      if (!response.ok) throw new Error('Failed to fetch recommendations');
      return response.json();
    },
    enabled: connectionData?.isConnected,
    refetchInterval: 5 * 60 * 1000, // Refresh every 5 minutes
  });

  // Mutation to connect Claude API
  const connectMutation = useMutation({
    mutationFn: async ({ apiKey, turnstileToken }: { apiKey: string; turnstileToken: string }) => {
      const response = await authApi.post('/api/claude-code/connect', {
        apiKey,
        turnstileToken,
      });
      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.message || 'Failed to connect Claude');
      }
      return response.json();
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['claude-connection'] });
      queryClient.invalidateQueries({ queryKey: ['trade-recommendations'] });
      setShowConnect(false);
      setApiKey('');
      setShowTurnstile(false);
    },
  });

  // Mutation to disconnect Claude
  const disconnectMutation = useMutation({
    mutationFn: async () => {
      const response = await authApi.delete('/api/claude-code/disconnect');
      if (!response.ok) throw new Error('Failed to disconnect');
      return response.json();
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['claude-connection'] });
      queryClient.invalidateQueries({ queryKey: ['trade-recommendations'] });
    },
  });

  // Mutation to analyze specific position
  const analyzePositionMutation = useMutation({
    mutationFn: async (positions: any[]) => {
      const response = await authApi.post('/api/claude-code/analyze-risk', { positions });
      if (!response.ok) throw new Error('Failed to analyze positions');
      return response.json();
    },
  });

  const handleConnect = () => {
    if (!apiKey) {
      alert('Please enter your Anthropic API key');
      return;
    }
    setShowTurnstile(true);
  };

  const handleTurnstileVerify = (token: string) => {
    setTurnstileToken(token);
    connectMutation.mutate({ apiKey, turnstileToken: token });
  };

  const formatRecommendation = (rec: TradeRecommendation) => {
    const profitLossRatio = rec.maxProfit / Math.abs(rec.maxLoss);
    return (
      <div className="border rounded-lg p-4 hover:shadow-lg transition-shadow">
        <div className="flex justify-between items-start mb-2">
          <h4 className="text-lg font-semibold">{rec.ticker}</h4>
          <span className={`px-2 py-1 text-xs rounded ${
            rec.pop >= 0.7 ? 'bg-green-100 text-green-800' : 'bg-yellow-100 text-yellow-800'
          }`}>
            {(rec.pop * 100).toFixed(0)}% POP
          </span>
        </div>
        
        <p className="text-sm font-medium text-gray-700 mb-1">{rec.strategy}</p>
        <p className="text-xs text-gray-600 mb-2">{rec.legs}</p>
        
        <p className="text-sm mb-3">{rec.thesis}</p>
        
        <div className="grid grid-cols-3 gap-2 text-xs">
          <div>
            <span className="text-gray-500">Max Loss:</span>
            <p className="font-medium text-red-600">${Math.abs(rec.maxLoss)}</p>
          </div>
          <div>
            <span className="text-gray-500">Max Profit:</span>
            <p className="font-medium text-green-600">${rec.maxProfit}</p>
          </div>
          <div>
            <span className="text-gray-500">Risk/Reward:</span>
            <p className="font-medium">1:{profitLossRatio.toFixed(2)}</p>
          </div>
        </div>
      </div>
    );
  };

  return (
    <div className="space-y-6">
      {/* Connection Status */}
      <div className="bg-white shadow rounded-lg p-6">
        <h2 className="text-xl font-semibold mb-4">AI Trading Assistant</h2>
        
        {isLoadingConnection ? (
          <div>Loading connection status...</div>
        ) : connectionData?.isConnected ? (
          <div>
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center space-x-2">
                <div className="w-3 h-3 bg-green-500 rounded-full"></div>
                <span className="text-green-700">Claude Connected</span>
              </div>
              <button
                onClick={() => disconnectMutation.mutate()}
                disabled={disconnectMutation.isPending}
                className="text-red-600 hover:text-red-800 text-sm"
              >
                Disconnect
              </button>
            </div>
            
            <p className="text-sm text-gray-600">
              Your AI assistant is ready to help with trade analysis and recommendations.
            </p>
          </div>
        ) : (
          <div>
            {!showConnect ? (
              <div>
                <p className="text-gray-600 mb-4">
                  Connect your Claude API to get AI-powered trading insights and recommendations.
                </p>
                <button
                  onClick={() => setShowConnect(true)}
                  className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700"
                >
                  Connect Claude Assistant
                </button>
              </div>
            ) : (
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Anthropic API Key
                  </label>
                  <input
                    type="password"
                    value={apiKey}
                    onChange={(e) => setApiKey(e.target.value)}
                    placeholder="sk-ant-api03-..."
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                  <p className="text-xs text-gray-500 mt-1">
                    Get your API key from the{' '}
                    <a href="https://console.anthropic.com/" target="_blank" rel="noopener noreferrer" className="text-blue-600 hover:underline">
                      Anthropic Console
                    </a>
                  </p>
                </div>
                
                {showTurnstile ? (
                  <div>
                    <p className="text-sm text-gray-600 mb-2">Please complete the verification:</p>
                    <TurnstileWidget
                      siteKey={turnstileSiteKey}
                      onVerify={handleTurnstileVerify}
                      onError={(error) => {
                        console.error('Turnstile error:', error);
                        alert('Verification failed. Please try again.');
                        setShowTurnstile(false);
                      }}
                      onExpire={() => {
                        setTurnstileToken('');
                        alert('Verification expired. Please try again.');
                        setShowTurnstile(false);
                      }}
                    />
                  </div>
                ) : (
                  <div className="flex space-x-2">
                    <button
                      onClick={handleConnect}
                      disabled={connectMutation.isPending}
                      className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700 disabled:opacity-50"
                    >
                      {connectMutation.isPending ? 'Connecting...' : 'Connect'}
                    </button>
                    <button
                      onClick={() => {
                        setShowConnect(false);
                        setApiKey('');
                      }}
                      className="px-4 py-2 border border-gray-300 rounded hover:bg-gray-50"
                    >
                      Cancel
                    </button>
                  </div>
                )}
              </div>
            )}
          </div>
        )}
      </div>

      {/* AI Trade Recommendations */}
      {connectionData?.isConnected && (
        <div className="bg-white shadow rounded-lg p-6">
          <div className="flex justify-between items-center mb-4">
            <h3 className="text-lg font-semibold">AI Trade Recommendations</h3>
            <button
              onClick={() => refetchRecommendations()}
              disabled={isLoadingRecs}
              className="text-blue-600 hover:text-blue-800 text-sm"
            >
              {isLoadingRecs ? 'Refreshing...' : 'Refresh'}
            </button>
          </div>
          
          {isLoadingRecs ? (
            <div className="text-center py-8">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
              <p className="text-gray-600 mt-4">Analyzing market data...</p>
            </div>
          ) : recommendations?.trades?.length > 0 ? (
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {recommendations.trades.map((rec: TradeRecommendation, index: number) => (
                <div key={index}>{formatRecommendation(rec)}</div>
              ))}
            </div>
          ) : (
            <p className="text-gray-500 text-center py-8">
              No trade recommendations available at this time.
            </p>
          )}
          
          {recommendations?.message && (
            <div className="mt-4 p-4 bg-yellow-50 border border-yellow-200 rounded">
              <p className="text-sm text-yellow-800">{recommendations.message}</p>
            </div>
          )}
        </div>
      )}

      {/* Risk Disclaimer */}
      <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
        <h4 className="font-semibold text-yellow-800 mb-2">AI Trading Disclaimer</h4>
        <p className="text-sm text-yellow-700">
          AI-generated recommendations are for educational purposes only and should not be considered financial advice. 
          Always perform your own due diligence and consult with a financial advisor before making trading decisions. 
          Past performance does not guarantee future results. Options trading involves substantial risk of loss.
        </p>
      </div>
    </div>
  );
}