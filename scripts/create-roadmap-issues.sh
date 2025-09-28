#!/bin/bash

# Script to create GitHub issues from ROADMAP.md
# Usage: ./scripts/create-roadmap-issues.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Creating GitHub Issues from ROADMAP.md${NC}"

# Check if gh CLI is installed
if ! command -v gh &> /dev/null; then
    echo -e "${RED}GitHub CLI (gh) is not installed. Please install it first.${NC}"
    echo "Visit: https://cli.github.com/"
    exit 1
fi

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo -e "${RED}Not in a git repository${NC}"
    exit 1
fi

# Function to create an issue
create_issue() {
    local id=$1
    local title=$2
    local phase=$3
    local milestone=$4
    local description=$5
    local labels=$6
    local complexity=$7
    
    echo -e "${YELLOW}Creating issue ${id}: ${title}${NC}"
    
    # Create issue body
    local body="## Task from YAML Independence Roadmap

**Task ID**: ${id}
**Phase**: ${phase}
**Milestone**: ${milestone}

### Description
${description}

### Acceptance Criteria
- [ ] Implementation complete
- [ ] Unit tests passing (>95% coverage)
- [ ] Documentation updated
- [ ] Code review approved
- [ ] Benchmarks show no regression

### References
- See [ROADMAP.md](../ROADMAP.md) for full context
- Architecture: [ARCHITECTURE_V2.md](../docs/ARCHITECTURE_V2.md)

---
*This issue is part of the YAML Library Independence v2.0 project*"

    # Create the issue
    gh issue create \
        --title "[${id}] ${title}" \
        --body "${body}" \
        --label "${labels}" || echo -e "${RED}Failed to create ${id}${NC}"
}

# Phase 1: Foundation Issues
echo -e "\n${GREEN}Phase 1: Foundation (Weeks 1-4)${NC}"

create_issue "YAML-001" "Implement byte stream reader with buffering" \
    "Phase 1: Foundation" "Milestone 1.1: Lexical Scanner" \
    "Create a byte stream reader that efficiently reads input with configurable buffering. Should handle UTF-8 encoding and track byte positions." \
    "roadmap,phase-1,component:scanner,type:feature" "M"

create_issue "YAML-002" "Create rune scanner with line/column tracking" \
    "Phase 1: Foundation" "Milestone 1.1: Lexical Scanner" \
    "Implement rune-level scanning with accurate line and column position tracking for error reporting." \
    "roadmap,phase-1,component:scanner,type:feature" "M"

create_issue "YAML-003" "Implement lookahead mechanism (n-character peek)" \
    "Phase 1: Foundation" "Milestone 1.1: Lexical Scanner" \
    "Add ability to peek ahead n characters without consuming them, essential for YAML parsing decisions." \
    "roadmap,phase-1,component:scanner,type:feature" "S"

create_issue "YAML-004" "Add position tracking and error reporting" \
    "Phase 1: Foundation" "Milestone 1.1: Lexical Scanner" \
    "Implement comprehensive position tracking (line, column, byte offset) and structured error reporting." \
    "roadmap,phase-1,component:scanner,type:feature" "M"

create_issue "YAML-005" "Create scanner configuration (tab width, encoding)" \
    "Phase 1: Foundation" "Milestone 1.1: Lexical Scanner" \
    "Add configurable scanner options including tab width, encoding settings, and buffer sizes." \
    "roadmap,phase-1,component:scanner,type:feature" "S"

create_issue "YAML-006" "Define token types enum" \
    "Phase 1: Foundation" "Milestone 1.2: Tokenizer" \
    "Create comprehensive token type definitions covering all YAML constructs using iota constants." \
    "roadmap,phase-1,component:tokenizer,type:feature" "S"

create_issue "YAML-007" "Implement scalar token recognition" \
    "Phase 1: Foundation" "Milestone 1.2: Tokenizer" \
    "Recognize and tokenize all scalar types: plain, single-quoted, double-quoted with escape sequences." \
    "roadmap,phase-1,component:tokenizer,type:feature" "L"

create_issue "YAML-008" "Implement structural tokens" \
    "Phase 1: Foundation" "Milestone 1.2: Tokenizer" \
    "Tokenize structural elements: document markers, key-value separators, sequence markers, flow indicators." \
    "roadmap,phase-1,component:tokenizer,type:feature" "L"

create_issue "YAML-009" "Implement special tokens" \
    "Phase 1: Foundation" "Milestone 1.2: Tokenizer" \
    "Handle comments, anchors, aliases, tags, and directives tokenization." \
    "roadmap,phase-1,component:tokenizer,type:feature" "M"

create_issue "YAML-010" "Add indentation tracking" \
    "Phase 1: Foundation" "Milestone 1.2: Tokenizer" \
    "Implement YAML indentation rules tracking for block structure parsing." \
    "roadmap,phase-1,component:tokenizer,type:feature" "M"

create_issue "YAML-011" "Set up YAML test suite from yaml-test-suite" \
    "Phase 1: Foundation" "Milestone 1.3: Testing Infrastructure" \
    "Integrate the official yaml-test-suite for compliance testing." \
    "roadmap,phase-1,component:testing,type:test" "M"

create_issue "YAML-012" "Create tokenizer test harness" \
    "Phase 1: Foundation" "Milestone 1.3: Testing Infrastructure" \
    "Build comprehensive test harness for tokenizer validation." \
    "roadmap,phase-1,component:testing,type:test" "M"

create_issue "YAML-013" "Add fuzzing tests for scanner" \
    "Phase 1: Foundation" "Milestone 1.3: Testing Infrastructure" \
    "Implement fuzz testing to find edge cases and potential panics." \
    "roadmap,phase-1,component:testing,type:test" "S"

create_issue "YAML-014" "Benchmark scanner performance" \
    "Phase 1: Foundation" "Milestone 1.3: Testing Infrastructure" \
    "Create performance benchmarks targeting >100MB/s parsing speed." \
    "roadmap,phase-1,component:testing,type:performance" "S"

# Phase 2: Parser Issues
echo -e "\n${GREEN}Phase 2: Parser (Weeks 5-8)${NC}"

create_issue "YAML-015" "Define event types interface" \
    "Phase 2: Parser" "Milestone 2.1: Event System" \
    "Design event-driven parser interface with comprehensive event types." \
    "roadmap,phase-2,component:parser,type:feature" "M"

create_issue "YAML-016" "Implement event stream" \
    "Phase 2: Parser" "Milestone 2.1: Event System" \
    "Create event stream mechanism for processing YAML documents." \
    "roadmap,phase-2,component:parser,type:feature" "L"

create_issue "YAML-017" "Create event emitter with observer pattern" \
    "Phase 2: Parser" "Milestone 2.1: Event System" \
    "Implement observer pattern for event-driven parsing." \
    "roadmap,phase-2,component:parser,type:feature" "M"

create_issue "YAML-018" "Add event filtering and transformation" \
    "Phase 2: Parser" "Milestone 2.1: Event System" \
    "Enable event filtering and transformation pipelines." \
    "roadmap,phase-2,component:parser,type:feature" "M"

create_issue "YAML-019" "Implement block scalar parser (literal |)" \
    "Phase 2: Parser" "Milestone 2.2: Block Style Parser" \
    "Parse literal block scalars with proper indentation handling." \
    "roadmap,phase-2,component:parser,type:feature" "L"

create_issue "YAML-020" "Implement folded scalar parser (>)" \
    "Phase 2: Parser" "Milestone 2.2: Block Style Parser" \
    "Parse folded block scalars with line folding rules." \
    "roadmap,phase-2,component:parser,type:feature" "L"

create_issue "YAML-021" "Add chomping indicators (+, -)" \
    "Phase 2: Parser" "Milestone 2.2: Block Style Parser" \
    "Handle block chomping indicators for trailing newlines." \
    "roadmap,phase-2,component:parser,type:feature" "S"

create_issue "YAML-022" "Handle indentation indicators" \
    "Phase 2: Parser" "Milestone 2.2: Block Style Parser" \
    "Process explicit indentation indicators in block scalars." \
    "roadmap,phase-2,component:parser,type:feature" "S"

create_issue "YAML-023" "Implement block sequences" \
    "Phase 2: Parser" "Milestone 2.2: Block Style Parser" \
    "Parse block-style sequences with proper nesting." \
    "roadmap,phase-2,component:parser,type:feature" "L"

create_issue "YAML-024" "Implement block mappings" \
    "Phase 2: Parser" "Milestone 2.2: Block Style Parser" \
    "Parse block-style mappings with complex keys support." \
    "roadmap,phase-2,component:parser,type:feature" "L"

create_issue "YAML-025" "Implement flow sequences []" \
    "Phase 2: Parser" "Milestone 2.3: Flow Style Parser" \
    "Parse flow sequences with proper comma handling." \
    "roadmap,phase-2,component:parser,type:feature" "M"

create_issue "YAML-026" "Implement flow mappings {}" \
    "Phase 2: Parser" "Milestone 2.3: Flow Style Parser" \
    "Parse flow mappings with key-value pairs." \
    "roadmap,phase-2,component:parser,type:feature" "M"

create_issue "YAML-027" "Handle flow style in block context" \
    "Phase 2: Parser" "Milestone 2.3: Flow Style Parser" \
    "Support flow collections within block structures." \
    "roadmap,phase-2,component:parser,type:feature" "M"

create_issue "YAML-028" "Add flow multiline support" \
    "Phase 2: Parser" "Milestone 2.3: Flow Style Parser" \
    "Handle multiline flow collections correctly." \
    "roadmap,phase-2,component:parser,type:feature" "S"

create_issue "YAML-029" "Handle multiple documents" \
    "Phase 2: Parser" "Milestone 2.4: Document Parser" \
    "Parse multiple YAML documents in a single stream." \
    "roadmap,phase-2,component:parser,type:feature" "M"

create_issue "YAML-030" "Process directives" \
    "Phase 2: Parser" "Milestone 2.4: Document Parser" \
    "Handle %YAML and %TAG directives properly." \
    "roadmap,phase-2,component:parser,type:feature" "M"

create_issue "YAML-031" "Implement explicit typing tags" \
    "Phase 2: Parser" "Milestone 2.4: Document Parser" \
    "Process explicit type tags (!!, !, custom tags)." \
    "roadmap,phase-2,component:parser,type:feature" "M"

create_issue "YAML-032" "Add merge key support (<<)" \
    "Phase 2: Parser" "Milestone 2.4: Document Parser" \
    "Implement merge key functionality for mapping merges." \
    "roadmap,phase-2,component:parser,type:feature" "M"

echo -e "\n${GREEN}Script complete! Created foundation and parser issues.${NC}"
echo -e "${YELLOW}Note: This is a partial script. Extend it to create all 106 issues.${NC}"
echo -e "${YELLOW}Run 'gh issue list --label roadmap' to see created issues.${NC}"
