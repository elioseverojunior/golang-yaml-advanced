# Architecture Documentation

## System Architecture

```mermaid
graph TB
    subgraph "Public API Layer"
        YAML[UnmarshalYAML]
        TOYAML[ToYAML]
        MERGE[MergeTrees]
        DIFF[DiffTrees]
        VALIDATE[Schema Validation]
        QUERY[Query System]
        TRANSFORM[Transform DSL]
        STREAM[Stream Parser]
    end

    subgraph "Core Data Structures"
        NT[NodeTree]
        DOC[Document]
        NODE[Node]
        ANCHOR[Anchors Map]
        DIRECTIVE[Directives]
    end

    subgraph "Node Types"
        DOCUMENT[DocumentNode]
        MAPPING[MappingNode]
        SEQUENCE[SequenceNode]
        SCALAR[ScalarNode]
        ALIAS[AliasNode]
    end

    subgraph "Metadata Preservation"
        COMMENTS[Comments<br/>- HeadComment<br/>- LineComment<br/>- FootComment]
        STYLE[Styles<br/>- Literal<br/>- Folded<br/>- Quoted<br/>- Flow]
        POSITION[Position<br/>- Line<br/>- Column]
        TAGS[Tags<br/>- !!str<br/>- !!int<br/>- Custom]
    end

    subgraph "Processing Engine"
        PARSER[YAML Parser<br/>yaml.v3]
        ENCODER[YAML Encoder<br/>2-space indent]
        MERGER[Merge Engine]
        DIFFER[Diff Engine]
        VALIDATOR[Validation Engine]
        TRANSFORMER[Transform Engine]
    end

    subgraph "Advanced Features"
        DSL[Transform DSL<br/>Fluent API]
        XPATH[XPath-like<br/>Queries]
        STREAMPROC[Stream<br/>Processor]
        SCHEMA[JSON Schema<br/>Validator]
    end

    YAML --> PARSER
    PARSER --> NT
    NT --> DOC
    DOC --> NODE
    NODE --> DOCUMENT
    NODE --> MAPPING
    NODE --> SEQUENCE
    NODE --> SCALAR
    NODE --> ALIAS

    NODE -.-> COMMENTS
    NODE -.-> STYLE
    NODE -.-> POSITION
    NODE -.-> TAGS
    NODE -.-> ANCHOR

    TOYAML --> ENCODER
    ENCODER --> NT

    MERGE --> MERGER
    MERGER --> NT

    DIFF --> DIFFER
    DIFFER --> NT

    VALIDATE --> VALIDATOR
    VALIDATOR --> SCHEMA

    QUERY --> XPATH
    TRANSFORM --> DSL
    STREAM --> STREAMPROC

    style YAML fill:#e1f5fe
    style TOYAML fill:#e1f5fe
    style MERGE fill:#e1f5fe
    style DIFF fill:#e1f5fe
    style VALIDATE fill:#e1f5fe
    style QUERY fill:#e1f5fe
    style TRANSFORM fill:#e1f5fe
    style STREAM fill:#e1f5fe

    style NT fill:#fff3e0
    style DOC fill:#fff3e0
    style NODE fill:#fff3e0

    style PARSER fill:#f3e5f5
    style ENCODER fill:#f3e5f5
    style MERGER fill:#f3e5f5
    style DIFFER fill:#f3e5f5
    style VALIDATOR fill:#f3e5f5
    style TRANSFORMER fill:#f3e5f5
```

## Application Flow - Parse and Merge

```mermaid
sequenceDiagram
    participant User
    participant API
    participant Parser
    participant NodeTree
    participant Merger
    participant Encoder
    participant Output

    User->>API: UnmarshalYAML(baseYAML)
    API->>Parser: Parse YAML content
    Parser->>Parser: Preserve comments
    Parser->>Parser: Preserve formatting
    Parser->>Parser: Convert integers (no sci notation)
    Parser->>NodeTree: Create base NodeTree
    NodeTree-->>API: Return base tree

    User->>API: UnmarshalYAML(overlayYAML)
    API->>Parser: Parse overlay content
    Parser->>NodeTree: Create overlay NodeTree
    NodeTree-->>API: Return overlay tree

    User->>API: MergeTrees(base, overlay)
    API->>Merger: Merge trees
    Merger->>Merger: Deep merge nodes
    Merger->>Merger: Preserve base comments
    Merger->>Merger: Apply overlay values
    Merger->>NodeTree: Create merged tree
    NodeTree-->>API: Return merged tree

    User->>API: ToYAML(mergedTree)
    API->>Encoder: Serialize tree
    Encoder->>Encoder: Apply 2-space indent
    Encoder->>Encoder: Preserve comments
    Encoder->>Encoder: Add empty lines
    Encoder->>Encoder: Format numbers
    Encoder->>Output: Generate YAML
    Output-->>User: Return YAML string
```

## Node Tree Structure

```mermaid
graph TD
    subgraph "NodeTree Structure"
        TREE[NodeTree]
        TREE --> DOC1[Document 1]
        TREE --> DOC2[Document 2]
        TREE --> DOCN[Document N]

        DOC1 --> ROOT1[Root Node<br/>Kind: MappingNode]

        ROOT1 --> KV1[Key-Value Pair 1]
        ROOT1 --> KV2[Key-Value Pair 2]
        ROOT1 --> KVN[Key-Value Pair N]

        KV1 --> KEY1[Key Node<br/>Kind: ScalarNode<br/>Value: 'name']
        KV1 --> VAL1[Value Node<br/>Kind: ScalarNode<br/>Value: 'MyApp']

        KV2 --> KEY2[Key Node<br/>Kind: ScalarNode<br/>Value: 'settings']
        KV2 --> VAL2[Value Node<br/>Kind: MappingNode]

        VAL2 --> NESTED1[Nested Key-Value 1]
        VAL2 --> NESTED2[Nested Key-Value 2]

        NESTED1 --> NKEY1[Key: 'debug']
        NESTED1 --> NVAL1[Value: true]

        NESTED2 --> NKEY2[Key: 'port']
        NESTED2 --> NVAL2[Value: 8080]
    end

    subgraph "Metadata Attached to Each Node"
        META[Node Metadata]
        META --> HC[HeadComment: string array]
        META --> LC[LineComment: string]
        META --> FC[FootComment: string array]
        META --> POS[Line: int<br/>Column: int]
        META --> ST[Style: NodeStyle]
        META --> TG[Tag: string]
        META --> ANC[Anchor: string]
    end

    VAL1 -.-> META

    style TREE fill:#e8f5e9
    style ROOT1 fill:#fff9c4
    style META fill:#fce4ec
```

## Transform DSL Flow

```mermaid
flowchart LR
    subgraph "Transform Pipeline"
        START[Input Tree] --> T1[Select Nodes]
        T1 --> T2[Filter Predicate]
        T2 --> T3[Map Function]
        T3 --> T4[Remove Keys]
        T4 --> T5[Rename Keys]
        T5 --> T6[Sort Keys]
        T6 --> T7[Add Comments]
        T7 --> T8[Flatten Structure]
        T8 --> END[Output Tree]
    end

    subgraph "Example Chain"
        EX1[NewTransformDSL] --> EX2[RemoveKey<br/>'password']
        EX2 --> EX3[RenameKey<br/>'user' → 'username']
        EX3 --> EX4[SortKeys]
        EX4 --> EX5[Apply<br/>to tree]
    end

    style START fill:#e1f5fe
    style END fill:#c8e6c9
```

## Validation Flow

```mermaid
flowchart TB
    subgraph "Schema Validation Process"
        INPUT[YAML Node] --> CHECK_TYPE{Type Check}
        CHECK_TYPE -->|Valid| CHECK_CONST[Constraint Check]
        CHECK_TYPE -->|Invalid| ERR1[Type Error]

        CHECK_CONST -->|Valid| CHECK_FORMAT[Format Check]
        CHECK_CONST -->|Invalid| ERR2[Constraint Error]

        CHECK_FORMAT -->|Valid| CHECK_ENUM[Enum Check]
        CHECK_FORMAT -->|Invalid| ERR3[Format Error]

        CHECK_ENUM -->|Valid| CHECK_COMPLEX[Complex Rules<br/>oneOf/anyOf/allOf]
        CHECK_ENUM -->|Invalid| ERR4[Enum Error]

        CHECK_COMPLEX -->|Valid| SUCCESS[Validation Success]
        CHECK_COMPLEX -->|Invalid| ERR5[Complex Rule Error]
    end

    subgraph "Supported Validations"
        TYPES[Types<br/>• string<br/>• number<br/>• boolean<br/>• object<br/>• array<br/>• null]

        FORMATS[Formats<br/>• email<br/>• uri<br/>• date<br/>• time<br/>• ipv4/ipv6<br/>• uuid]

        CONSTRAINTS[Constraints<br/>• minLength<br/>• maxLength<br/>• minimum<br/>• maximum<br/>• pattern<br/>• required]
    end

    style SUCCESS fill:#c8e6c9
    style ERR1 fill:#ffcdd2
    style ERR2 fill:#ffcdd2
    style ERR3 fill:#ffcdd2
    style ERR4 fill:#ffcdd2
    style ERR5 fill:#ffcdd2
```

## Stream Parser Architecture

```mermaid
flowchart TD
    subgraph "Stream Processing"
        FILE[Large YAML File<br/>>100MB] --> READER[Buffered Reader]
        READER --> SCANNER[Document Scanner]
        SCANNER --> BOUNDARY{Document<br/>Boundary?}

        BOUNDARY -->|Yes| PARSE[Parse Document]
        BOUNDARY -->|No| CONTINUE[Continue Reading]

        PARSE --> CALLBACK[Document Callback]
        CALLBACK --> PROCESS{Process<br/>Document?}

        PROCESS -->|Yes| HANDLE[Handle Document]
        PROCESS -->|No| SKIP[Skip Document]

        HANDLE --> MORE{More<br/>Documents?}
        SKIP --> MORE

        MORE -->|Yes| SCANNER
        MORE -->|No| COMPLETE[Complete]

        CONTINUE --> SCANNER
    end

    subgraph "Memory Management"
        MEM1[Document 1<br/>In Memory]
        MEM2[Document 2<br/>Released]
        MEM3[Document N<br/>In Memory]

        MEM1 -.-> GC[Garbage<br/>Collection]
        MEM2 -.-> GC
    end

    HANDLE -.-> MEM1

    style FILE fill:#fff3e0
    style COMPLETE fill:#c8e6c9
```

## Merge Strategy

```mermaid
flowchart TB
    subgraph "Merge Decision Tree"
        START[Two Nodes to Merge] --> TYPE{Node Types<br/>Match?}

        TYPE -->|No| OVERLAY[Use Overlay Node]
        TYPE -->|Yes| KIND{What Kind?}

        KIND -->|Scalar| REPLACE[Replace with Overlay]
        KIND -->|Sequence| CONCAT[Concatenate Arrays]
        KIND -->|Mapping| DEEP[Deep Merge]

        DEEP --> ITER[Iterate Keys]
        ITER --> EXISTS{Key Exists<br/>in Base?}

        EXISTS -->|No| ADD[Add Key-Value]
        EXISTS -->|Yes| RECURSE[Merge Values<br/>Recursively]

        RECURSE --> TYPE

        ADD --> NEXT{More Keys?}
        RECURSE --> NEXT

        NEXT -->|Yes| ITER
        NEXT -->|No| PRESERVE[Preserve Comments<br/>From Base]

        PRESERVE --> RESULT[Merged Node]
        OVERLAY --> RESULT
        REPLACE --> RESULT
        CONCAT --> RESULT
    end

    style START fill:#e1f5fe
    style RESULT fill:#c8e6c9
```

## Query System

```mermaid
flowchart LR
    subgraph "Query Path Resolution"
        QUERY[Query String<br/>'app.settings.port'] --> SPLIT[Split by '.']
        SPLIT --> PARTS[Path Parts<br/>'app', 'settings', 'port']
        PARTS --> TRAVERSE[Traverse Tree]

        TRAVERSE --> NODE1[Find 'app' Node]
        NODE1 --> NODE2[Find 'settings' Child]
        NODE2 --> NODE3[Find 'port' Child]
        NODE3 --> RESULT[Return Value<br/>8080]
    end

    subgraph "Advanced Queries"
        WILD[Wildcard Query<br/>'*.name'] --> ALL[Find All 'name' Keys]
        ARRAY[Array Query<br/>'items at index 0'] --> INDEX[Access by Index]
        LAST[Last Item<br/>'items at index -1'] --> REVERSE[Reverse Index]
    end

    style QUERY fill:#e1f5fe
    style RESULT fill:#c8e6c9
```