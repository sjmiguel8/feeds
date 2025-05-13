import yaml
import os

class IncludeLoader(yaml.SafeLoader):
    """YAML Loader that includes files referenced by !include tags."""
    def __init__(self, stream, base_path):
        super().__init__(stream)
        self.base_path = base_path

    def include(self, node):
        filename = os.path.join(self.base_path, self.construct_scalar(node))
        with open(filename, 'r') as f:
            nested_base_path = os.path.dirname(filename)
            return load_config(filename)

IncludeLoader.add_constructor('!include', IncludeLoader.include)

def load_config(config_path):
    """Load a YAML configuration file using the IncludeLoader."""
    base_path = os.path.dirname(config_path)
    with open(config_path, 'r') as f:
        return yaml.load(f, Loader=IncludeLoader, base_path=base_path)

if __name__ == '__main__':
    # Example usage:
    config = load_config('/Users/rishyscott/Documents/coding/glancem/glance/config/glance.yml')
    print(config)

