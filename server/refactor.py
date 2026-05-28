import os
import re
import glob

files = glob.glob("internal/handlers/*.go")

for file in files:
    with open(file, 'r') as f:
        content = f.read()

    orig = content

    # Add import if we need to replace
    if "utils." not in content and "return c.Status" in content:
        content = re.sub(r'import\s*\((.*?)\)', r'import (\1\n\t"chantry/server/internal/utils"\n)', content, count=1, flags=re.DOTALL)

    # Match error maps with "error" or "message"
    # Matches: return c.Status(CODE).JSON(fiber.Map{ "error": MSG })
    # We use non-greedy matching for the inside of the map, ensuring it only has one key-value pair.
    content = re.sub(
        r'return\s+c\.Status\(([^)]+)\)\.JSON\(fiber\.Map\{\s*"(?:error|message)"\s*:\s*([^,}]+),?\s*\}\)',
        r'return utils.JSONError(c, \1, \2)',
        content,
        flags=re.DOTALL
    )

    # Match JSONSuccess with a variable/struct (no fiber.Map)
    # Match: return c.Status(CODE).JSON(VAR) where VAR doesn't contain '{'
    content = re.sub(
        r'return\s+c\.Status\(([^)]+)\)\.JSON\(([^(){\n]+)\)',
        r'return utils.JSONSuccess(c, \1, \2)',
        content
    )

    # Match JSONSuccess with fiber.Map{} (we'll just wrap the whole fiber.Map in JSONSuccess)
    # Actually, if it's a success map, it should go to data: fiber.Map{...}
    # return c.Status(CODE).JSON(fiber.Map{ ... })
    content = re.sub(
        r'return\s+c\.Status\(([^)]+)\)\.JSON\((fiber\.Map\{.*?\})\)',
        r'return utils.JSONSuccess(c, \1, \2)',
        content,
        flags=re.DOTALL
    )

    # Wait, the previous DOTALL might be dangerous. Let's do it line by line or with a safer regex.
    
    if content != orig:
        with open(file, 'w') as f:
            f.write(content)

# Find remaining c.Status manually
os.system("grep -n 'c.Status' internal/handlers/*.go")
