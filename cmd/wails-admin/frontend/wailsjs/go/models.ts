export namespace main {
	
	export class AuditLogEntry {
	    time: string;
	    type: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new AuditLogEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.time = source["time"];
	        this.type = source["type"];
	        this.message = source["message"];
	    }
	}
	export class DashboardStats {
	    status: string;
	    uptime: string;
	    active_php_workers: number;
	    memory_allocated_mb: number;
	    memory_sys_mb: number;
	    total_sites: number;
	    api_port: number;
	
	    static createFrom(source: any = {}) {
	        return new DashboardStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.uptime = source["uptime"];
	        this.active_php_workers = source["active_php_workers"];
	        this.memory_allocated_mb = source["memory_allocated_mb"];
	        this.memory_sys_mb = source["memory_sys_mb"];
	        this.total_sites = source["total_sites"];
	        this.api_port = source["api_port"];
	    }
	}
	export class PHPExtension {
	    name: string;
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PHPExtension(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.enabled = source["enabled"];
	    }
	}
	export class PHPTuningConfig {
	    opcache_enabled: boolean;
	    memory_limit: string;
	    upload_max_filesize: string;
	    post_max_size: string;
	    max_execution_time: string;
	    opcache_memory: string;
	
	    static createFrom(source: any = {}) {
	        return new PHPTuningConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.opcache_enabled = source["opcache_enabled"];
	        this.memory_limit = source["memory_limit"];
	        this.upload_max_filesize = source["upload_max_filesize"];
	        this.post_max_size = source["post_max_size"];
	        this.max_execution_time = source["max_execution_time"];
	        this.opcache_memory = source["opcache_memory"];
	    }
	}
	export class PHPVersionInfo {
	    version: string;
	    path: string;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new PHPVersionInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.path = source["path"];
	        this.status = source["status"];
	    }
	}
	export class ServerConfig {
	    api_port: number;
	    data_dir: string;
	    php_workers: number;
	    api_key: string;
	
	    static createFrom(source: any = {}) {
	        return new ServerConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.api_port = source["api_port"];
	        this.data_dir = source["data_dir"];
	        this.php_workers = source["php_workers"];
	        this.api_key = source["api_key"];
	    }
	}
	export class SiteConfig {
	    name: string;
	    domain: string;
	    port: number;
	    document_root: string;
	    php_version: string;
	
	    static createFrom(source: any = {}) {
	        return new SiteConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.domain = source["domain"];
	        this.port = source["port"];
	        this.document_root = source["document_root"];
	        this.php_version = source["php_version"];
	    }
	}
	export class SitesFile {
	    sites: SiteConfig[];
	
	    static createFrom(source: any = {}) {
	        return new SitesFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sites = this.convertValues(source["sites"], SiteConfig);
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

