import { useState } from 'react';
import './App.css';
import { ReactFlowProvider } from '@xyflow/react';
import { useWebSocket } from './hooks/useWebSocket';
import { useAgentGraph } from './hooks/useAgentGraph';
import { PromptInput } from './components/PromptInput';
import { EventFeed } from './components/EventFeed';
import { AgentGraph } from './components/AgentGraph';

function App() {
  const { events, isConnected, isRunning, sendMessage, clearEvents } = useWebSocket();
  const { nodes, edges } = useAgentGraph(events);
  const [feedVisible, setFeedVisible] = useState(true);

  return (
    <div className="app">
      <header className="app-header">
        <h1>AI Engine</h1>
        <div className="connection-status">
          <span className={`status-dot ${isConnected ? 'connected' : 'disconnected'}`} />
          <span>{isConnected ? 'Connected' : 'Disconnected'}</span>
        </div>
        {!isRunning && events.length > 0 && (
          <button className="clear-btn" onClick={clearEvents}>
            Clear
          </button>
        )}
        <button
          className="toggle-feed-btn"
          onClick={() => setFeedVisible((v) => !v)}
          title={feedVisible ? 'Hide event feed' : 'Show event feed'}
        >
          {feedVisible ? 'Hide Feed' : 'Show Feed'}
        </button>
      </header>

      <main className="app-main">
        <ReactFlowProvider>
          <AgentGraph nodes={nodes} edges={edges} />
        </ReactFlowProvider>
        <div className={`feed-panel${feedVisible ? '' : ' feed-panel--hidden'}`}>
          <EventFeed events={events} />
        </div>
      </main>

      <footer className="app-footer">
        <PromptInput onSend={sendMessage} disabled={isRunning} />
      </footer>
    </div>
  );
}

export default App;
