import { useState, useEffect } from 'react';
import { CheckCircle, AlertCircle, Loader2 } from 'lucide-react';
import { GetSimilarityStatus } from '../../wailsjs/go/main/App';
import clsx from 'clsx';

const AIStatusIndicator = ({ className, showLabel = false }) => {
  const [status, setStatus] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const checkStatus = async () => {
      try {
        const result = await GetSimilarityStatus();
        setStatus(result);
      } catch (err) {
        setStatus({ available: false });
      } finally {
        setLoading(false);
      }
    };

    checkStatus();
    const interval = setInterval(checkStatus, 30000); // Check every 30s
    return () => clearInterval(interval);
  }, []);

  if (loading) {
    return <Loader2 className={clsx("animate-spin text-muted", className)} size={14} />;
  }

  if (!status) {
    return null;
  }

  if (status.available) {
    return (
      <div className="flex items-center gap-1">
        <CheckCircle className={clsx("text-green-500", className)} size={14} />
        {showLabel && <span className="text-xs text-muted">AI Ready</span>}
      </div>
    );
  }

  return (
    <div className="flex items-center gap-1">
      <AlertCircle className={clsx("text-orange-500", className)} size={14} />
      {showLabel && <span className="text-xs text-muted">AI Offline</span>}
    </div>
  );
};

export default AIStatusIndicator;
