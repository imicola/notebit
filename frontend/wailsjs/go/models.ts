export namespace config {
	
	export class ChunkingConfig {
	    strategy: string;
	    chunk_size: number;
	    chunk_overlap: number;
	    min_chunk_size: number;
	    max_chunk_size: number;
	    preserve_heading: boolean;
	    heading_separator: string;
	
	    static createFrom(source: any = {}) {
	        return new ChunkingConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.strategy = source["strategy"];
	        this.chunk_size = source["chunk_size"];
	        this.chunk_overlap = source["chunk_overlap"];
	        this.min_chunk_size = source["min_chunk_size"];
	        this.max_chunk_size = source["max_chunk_size"];
	        this.preserve_heading = source["preserve_heading"];
	        this.heading_separator = source["heading_separator"];
	    }
	}
	export class GraphConfig {
	    min_similarity_threshold: number;
	    max_nodes: number;
	    show_implicit_links: boolean;
	
	    static createFrom(source: any = {}) {
	        return new GraphConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.min_similarity_threshold = source["min_similarity_threshold"];
	        this.max_nodes = source["max_nodes"];
	        this.show_implicit_links = source["show_implicit_links"];
	    }
	}
	export class OllamaConfig {
	    base_url: string;
	    embedding_model: string;
	    timeout: number;
	
	    static createFrom(source: any = {}) {
	        return new OllamaConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.base_url = source["base_url"];
	        this.embedding_model = source["embedding_model"];
	        this.timeout = source["timeout"];
	    }
	}
	export class OpenAIConfig {
	    api_key: string;
	    base_url: string;
	    organization: string;
	    embedding_model: string;
	
	    static createFrom(source: any = {}) {
	        return new OpenAIConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.api_key = source["api_key"];
	        this.base_url = source["base_url"];
	        this.organization = source["organization"];
	        this.embedding_model = source["embedding_model"];
	    }
	}
	export class LLMConfig {
	    provider: string;
	    model: string;
	    temperature: number;
	    max_tokens: number;
	    openai: OpenAIConfig;
	    ollama: OllamaConfig;
	
	    static createFrom(source: any = {}) {
	        return new LLMConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.provider = source["provider"];
	        this.model = source["model"];
	        this.temperature = source["temperature"];
	        this.max_tokens = source["max_tokens"];
	        this.openai = this.convertValues(source["openai"], OpenAIConfig);
	        this.ollama = this.convertValues(source["ollama"], OllamaConfig);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	export class RAGConfig {
	    max_context_chunks: number;
	    temperature: number;
	    system_prompt: string;
	
	    static createFrom(source: any = {}) {
	        return new RAGConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.max_context_chunks = source["max_context_chunks"];
	        this.temperature = source["temperature"];
	        this.system_prompt = source["system_prompt"];
	    }
	}

}

export namespace database {
	
	export class Chunk {
	    id: number;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	    file_id: number;
	    content: string;
	    heading: string;
	    embedding: number[];
	    embedding_model: string;
	    // Go type: time
	    embedding_created_at?: any;
	    vec_indexed: boolean;
	    embedding_dim?: number;
	
	    static createFrom(source: any = {}) {
	        return new Chunk(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	        this.file_id = source["file_id"];
	        this.content = source["content"];
	        this.heading = source["heading"];
	        this.embedding = source["embedding"];
	        this.embedding_model = source["embedding_model"];
	        this.embedding_created_at = this.convertValues(source["embedding_created_at"], null);
	        this.vec_indexed = source["vec_indexed"];
	        this.embedding_dim = source["embedding_dim"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Tag {
	    id: number;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	    name: string;
	    color: string;
	
	    static createFrom(source: any = {}) {
	        return new Tag(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	        this.name = source["name"];
	        this.color = source["color"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class File {
	    id: number;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	    path: string;
	    title: string;
	    content_hash: string;
	    last_modified: number;
	    file_size: number;
	    chunks?: Chunk[];
	    tags?: Tag[];
	
	    static createFrom(source: any = {}) {
	        return new File(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	        this.path = source["path"];
	        this.title = source["title"];
	        this.content_hash = source["content_hash"];
	        this.last_modified = source["last_modified"];
	        this.file_size = source["file_size"];
	        this.chunks = this.convertValues(source["chunks"], Chunk);
	        this.tags = this.convertValues(source["tags"], Tag);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace files {
	
	export class JSONTime {
	
	
	    static createFrom(source: any = {}) {
	        return new JSONTime(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class FileNode {
	    name: string;
	    path: string;
	    isDir: boolean;
	    modifiedTime: JSONTime;
	    size: number;
	    children?: FileNode[];
	
	    static createFrom(source: any = {}) {
	        return new FileNode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.isDir = source["isDir"];
	        this.modifiedTime = this.convertValues(source["modifiedTime"], JSONTime);
	        this.size = source["size"];
	        this.children = this.convertValues(source["children"], FileNode);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class NoteContent {
	    path: string;
	    content: string;
	
	    static createFrom(source: any = {}) {
	        return new NoteContent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.content = source["content"];
	    }
	}

}

export namespace gorm {
	
	export class DeletedAt {
	    // Go type: time
	    Time: any;
	    Valid: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DeletedAt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Time = this.convertValues(source["Time"], null);
	        this.Valid = source["Valid"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace graph {
	
	export class Link {
	    source: string;
	    target: string;
	    type: string;
	    strength: number;
	
	    static createFrom(source: any = {}) {
	        return new Link(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.source = source["source"];
	        this.target = source["target"];
	        this.type = source["type"];
	        this.strength = source["strength"];
	    }
	}
	export class Node {
	    id: string;
	    label: string;
	    type: string;
	    path: string;
	    size: number;
	    val: number;
	
	    static createFrom(source: any = {}) {
	        return new Node(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.label = source["label"];
	        this.type = source["type"];
	        this.path = source["path"];
	        this.size = source["size"];
	        this.val = source["val"];
	    }
	}
	export class GraphData {
	    nodes: Node[];
	    links: Link[];
	
	    static createFrom(source: any = {}) {
	        return new GraphData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.nodes = this.convertValues(source["nodes"], Node);
	        this.links = this.convertValues(source["links"], Link);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	

}

export namespace main {
	
	export class SimilarNote {
	    path: string;
	    title: string;
	    content: string;
	    heading: string;
	    similarity: number;
	    chunk_id: number;
	
	    static createFrom(source: any = {}) {
	        return new SimilarNote(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.title = source["title"];
	        this.content = source["content"];
	        this.heading = source["heading"];
	        this.similarity = source["similarity"];
	        this.chunk_id = source["chunk_id"];
	    }
	}

}

