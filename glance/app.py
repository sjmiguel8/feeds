# ... existing imports ...
from glance.config_loader import load_config

# Load the configuration
config = load_config('/Users/rishyscott/Documents/coding/glancem/glance/config/glance.yml')

# ... use the config ...
print(config['server']['assets-path'])
