import { query, SDKMessage } from '@anthropic-ai/claude-code';
import { spawn } from 'child_process';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

export class ClaudeHandler {
  constructor(logger) {
    this.logger = logger;
    this.sessions = new Map();
    this.systemPrompt = `You are Claude, an AI trading assistant integrated with VibeTrade. You have access to:

1. Real-time market data and options chains
2. Portfolio management capabilities
3. Trading strategy analysis
4. Risk assessment tools

You can help users with:
- Analyzing market conditions and finding trading opportunities
- Evaluating options strategies (spreads, condors, strangles, etc.)
- Risk management and portfolio optimization
- Educational explanations of trading concepts
- Creating custom trading scripts and strategies

Always prioritize risk management and provide clear explanations of potential outcomes. When discussing trades, include:
- Maximum profit/loss scenarios
- Probability of profit
- Key risks to consider
- Market conditions that would affect the trade

Be helpful, accurate, and educational in your responses.`;
  }

  async processMessage(sessionId, content, options = {}) {
    try {
      // Get or create session
      let session = this.sessions.get(sessionId);
      if (!session) {
        session = {
          id: sessionId,
          messages: [],
          context: {},
        };
        this.sessions.set(sessionId, session);
      }

      // Add user message to history
      session.messages.push({
        role: 'user',
        content: content,
        timestamp: new Date(),
      });

      // Prepare the prompt with trading context
      const enhancedPrompt = this.enhancePromptWithTradingContext(content, session);

      // Call Claude Code SDK
      this.logger.info(`Processing message for session ${sessionId}`);
      
      const messages = [];
      const abortController = new AbortController();

      try {
        for await (const message of query({
          prompt: enhancedPrompt,
          abortController,
          options: {
            maxTurns: 3,
            systemPrompt: this.systemPrompt,
            allowedTools: ['Read', 'WebSearch', 'WebFetch'],
            cwd: process.cwd(),
          },
        })) {
          messages.push(message);
          
          // Handle streaming if callback provided
          if (options.onStream && message.type === 'assistant') {
            const content = this.extractContentFromMessage(message);
            if (content) {
              options.onStream(content);
            }
          }
        }
      } catch (error) {
        this.logger.error('Error calling Claude Code SDK:', error);
        throw error;
      }

      // Extract the final response
      const response = this.extractFinalResponse(messages);
      
      // Add assistant response to history
      session.messages.push({
        role: 'assistant',
        content: response,
        timestamp: new Date(),
      });

      return response;
    } catch (error) {
      this.logger.error('Error processing message:', error);
      throw error;
    }
  }

  enhancePromptWithTradingContext(prompt, session) {
    // Add trading-specific context to the prompt
    let enhanced = prompt;

    // Add market context if available
    if (session.context.currentMarketData) {
      enhanced += `\n\nCurrent market context: ${JSON.stringify(session.context.currentMarketData)}`;
    }

    // Add portfolio context if available
    if (session.context.portfolio) {
      enhanced += `\n\nPortfolio: ${JSON.stringify(session.context.portfolio)}`;
    }

    return enhanced;
  }

  extractContentFromMessage(message) {
    if (message.message?.content) {
      if (typeof message.message.content === 'string') {
        return message.message.content;
      } else if (Array.isArray(message.message.content)) {
        return message.message.content
          .filter(item => item.type === 'text')
          .map(item => item.text)
          .join('');
      }
    }
    return '';
  }

  extractFinalResponse(messages) {
    // Find the last assistant message or result
    for (let i = messages.length - 1; i >= 0; i--) {
      const message = messages[i];
      
      if (message.type === 'result' && message.result) {
        return message.result;
      } else if (message.type === 'assistant') {
        const content = this.extractContentFromMessage(message);
        if (content) {
          return content;
        }
      }
    }

    return 'I apologize, but I couldn\'t generate a response. Please try again.';
  }

  clearSession(sessionId) {
    this.sessions.delete(sessionId);
    this.logger.info(`Cleared session ${sessionId}`);
  }

  cleanupSession(sessionId) {
    // Keep session for a while in case of reconnection
    setTimeout(() => {
      if (this.sessions.has(sessionId)) {
        this.sessions.delete(sessionId);
        this.logger.info(`Cleaned up session ${sessionId}`);
      }
    }, 300000); // 5 minutes
  }

  getSessionHistory(sessionId) {
    const session = this.sessions.get(sessionId);
    return session ? session.messages : [];
  }

  updateSessionContext(sessionId, context) {
    const session = this.sessions.get(sessionId);
    if (session) {
      session.context = { ...session.context, ...context };
    }
  }
}