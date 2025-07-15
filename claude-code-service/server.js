import express from 'express';
import { WebSocketServer } from 'ws';
import http from 'http';
import dotenv from 'dotenv';
import winston from 'winston';
import { ClaudeHandler } from './claude-handler.js';

// Load environment variables
dotenv.config();

// Configure logger
const logger = winston.createLogger({
  level: 'info',
  format: winston.format.json(),
  transports: [
    new winston.transports.Console({
      format: winston.format.simple(),
    }),
  ],
});

// Create Express app
const app = express();
const port = process.env.PORT || 3001;

// Middleware
app.use(express.json());

// Health check endpoint
app.get('/health', (req, res) => {
  res.json({ status: 'ok', timestamp: new Date().toISOString() });
});

// Create HTTP server
const server = http.createServer(app);

// Create WebSocket server
const wss = new WebSocketServer({ 
  server,
  path: '/ws',
});

// Initialize Claude handler
const claudeHandler = new ClaudeHandler(logger);

// WebSocket connection handler
wss.on('connection', (ws, req) => {
  const clientId = Date.now().toString();
  logger.info(`New WebSocket connection: ${clientId}`);

  // Send initial connection message
  ws.send(JSON.stringify({
    type: 'system',
    subtype: 'connected',
    session_id: clientId,
    message: 'Connected to Claude Code service',
  }));

  // Handle incoming messages
  ws.on('message', async (data) => {
    try {
      const message = JSON.parse(data.toString());
      logger.info(`Received message from ${clientId}:`, message);

      // Handle different message types
      switch (message.type) {
        case 'chat':
          await handleChatMessage(ws, clientId, message);
          break;
        
        case 'command':
          await handleCommand(ws, clientId, message);
          break;
        
        default:
          ws.send(JSON.stringify({
            type: 'error',
            message: `Unknown message type: ${message.type}`,
          }));
      }
    } catch (error) {
      logger.error('Error processing message:', error);
      ws.send(JSON.stringify({
        type: 'error',
        message: 'Failed to process message',
        error: error.message,
      }));
    }
  });

  // Handle disconnection
  ws.on('close', () => {
    logger.info(`WebSocket disconnected: ${clientId}`);
    claudeHandler.cleanupSession(clientId);
  });

  // Handle errors
  ws.on('error', (error) => {
    logger.error(`WebSocket error for ${clientId}:`, error);
  });
});

// Handle chat messages
async function handleChatMessage(ws, sessionId, message) {
  try {
    // Send typing indicator
    ws.send(JSON.stringify({
      type: 'system',
      subtype: 'typing',
      session_id: sessionId,
    }));

    // Process message with Claude
    const response = await claudeHandler.processMessage(sessionId, message.content, {
      onStream: (chunk) => {
        // Send streaming response
        ws.send(JSON.stringify({
          type: 'assistant',
          subtype: 'stream',
          content: chunk,
          session_id: sessionId,
          isStreaming: true,
        }));
      },
    });

    // Send final response
    ws.send(JSON.stringify({
      type: 'assistant',
      content: response,
      session_id: sessionId,
      timestamp: new Date().toISOString(),
    }));
  } catch (error) {
    logger.error('Error in chat handler:', error);
    ws.send(JSON.stringify({
      type: 'error',
      message: 'Failed to process chat message',
      error: error.message,
    }));
  }
}

// Handle commands (e.g., clear, export, settings)
async function handleCommand(ws, sessionId, message) {
  try {
    switch (message.command) {
      case 'clear':
        claudeHandler.clearSession(sessionId);
        ws.send(JSON.stringify({
          type: 'system',
          subtype: 'cleared',
          message: 'Session cleared',
          session_id: sessionId,
        }));
        break;

      case 'export':
        const history = claudeHandler.getSessionHistory(sessionId);
        ws.send(JSON.stringify({
          type: 'system',
          subtype: 'export',
          data: history,
          session_id: sessionId,
        }));
        break;

      default:
        ws.send(JSON.stringify({
          type: 'error',
          message: `Unknown command: ${message.command}`,
        }));
    }
  } catch (error) {
    logger.error('Error in command handler:', error);
    ws.send(JSON.stringify({
      type: 'error',
      message: 'Failed to process command',
      error: error.message,
    }));
  }
}

// Start server
server.listen(port, () => {
  logger.info(`Claude Code service running on port ${port}`);
  logger.info(`WebSocket endpoint: ws://localhost:${port}/ws`);
});