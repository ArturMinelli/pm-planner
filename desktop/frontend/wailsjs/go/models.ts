export namespace app {
	
	export class PlannerPayload {
	    date: string;
	    targetSecs: number;
	    originalTimes: string[];
	    in1: string;
	    out1: string;
	    in2: string;
	    out2: string;
	    originalsLine: string;
	
	    static createFrom(source: any = {}) {
	        return new PlannerPayload(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.date = source["date"];
	        this.targetSecs = source["targetSecs"];
	        this.originalTimes = source["originalTimes"];
	        this.in1 = source["in1"];
	        this.out1 = source["out1"];
	        this.in2 = source["in2"];
	        this.out2 = source["out2"];
	        this.originalsLine = source["originalsLine"];
	    }
	}
	export class PlannerSummary {
	    out2: string;
	    firstSpanSecs: number;
	    secondSpanSecs: number;
	    totalSpanSecs: number;
	    overtimeSecs: number;
	
	    static createFrom(source: any = {}) {
	        return new PlannerSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.out2 = source["out2"];
	        this.firstSpanSecs = source["firstSpanSecs"];
	        this.secondSpanSecs = source["secondSpanSecs"];
	        this.totalSpanSecs = source["totalSpanSecs"];
	        this.overtimeSecs = source["overtimeSecs"];
	    }
	}

}

export namespace config {
	
	export class File {
	    email: string;
	    password: string;
	    cache_ttl_hours?: number;
	
	    static createFrom(source: any = {}) {
	        return new File(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.email = source["email"];
	        this.password = source["password"];
	        this.cache_ttl_hours = source["cache_ttl_hours"];
	    }
	}

}

