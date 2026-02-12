import React from 'react';
import { AlertTriangle, RefreshCw } from 'lucide-react';
import { ERRORS } from '../constants';

class ErrorBoundary extends React.Component {
  constructor(props) {
    super(props);
    this.state = { 
      hasError: false, 
      error: null,
      errorInfo: null
    };
  }

  static getDerivedStateFromError(error) {
    // Update state so the next render will show the fallback UI.
    return { hasError: true, error };
  }

  componentDidCatch(error, errorInfo) {
    // You can also log the error to an error reporting service
    console.error("Uncaught error:", error, errorInfo);
    this.setState({ errorInfo });
  }

  handleReload = () => {
    window.location.reload();
  };

  render() {
    if (this.state.hasError) {
      // Fallback UI
      return (
        <div className="flex flex-col items-center justify-center h-screen w-full bg-primary text-normal p-6">
          <div className="flex flex-col items-center max-w-md text-center p-8 bg-secondary rounded-lg border border-modifier-border shadow-lg">
            <div className="text-obsidian-red mb-4">
              <AlertTriangle size={48} />
            </div>
            
            <h2 className="text-xl font-bold mb-2 text-normal">
              {ERRORS.APP_CRASH || 'Something went wrong'}
            </h2>
            
            <p className="text-muted mb-6 text-sm">
              We apologize for the inconvenience. An unexpected error has occurred.
            </p>

            {this.state.error && (
              <div className="w-full mb-6 p-3 bg-primary rounded border border-modifier-border overflow-auto max-h-40 text-left">
                <code className="text-xs text-obsidian-red font-mono break-all">
                  {this.state.error.toString()}
                </code>
              </div>
            )}

            <button
              onClick={this.handleReload}
              className="flex items-center gap-2 px-4 py-2 bg-modifier-hover text-normal border border-modifier-border rounded hover:bg-modifier-border-focus transition-colors"
            >
              <RefreshCw size={16} />
              <span>Reload Application</span>
            </button>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}

export default ErrorBoundary;
