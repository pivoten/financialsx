export namespace auth {
	
	export class User {
	    id: number;
	    username: string;
	    email?: string;
	    company_name: string;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    last_login?: any;
	
	    static createFrom(source: any = {}) {
	        return new User(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.username = source["username"];
	        this.email = source["email"];
	        this.company_name = source["company_name"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.last_login = this.convertValues(source["last_login"], null);
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

export namespace company {
	
	export class Company {
	    name: string;
	    data_path: string;
	    has_dbf: boolean;
	    has_sql: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Company(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.data_path = source["data_path"];
	        this.has_dbf = source["has_dbf"];
	        this.has_sql = source["has_sql"];
	    }
	}

}

