export namespace cleanup {
	
	export class Result {
	    moved: number;
	    deleted: number;
	    errors?: string[];
	
	    static createFrom(source: any = {}) {
	        return new Result(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.moved = source["moved"];
	        this.deleted = source["deleted"];
	        this.errors = source["errors"];
	    }
	}

}

export namespace queue {
	
	export class Job {
	    id: string;
	    path: string;
	    filename: string;
	    status: string;
	    progress: number;
	    elapsed: string;
	    fps: number;
	    originalSize: number;
	    outputSize: number;
	    savings: number;
	    error?: string;
	    outputPath?: string;
	    skipReason?: string;
	    originalDeleted?: boolean;
	    // Go type: time
	    addedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Job(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.path = source["path"];
	        this.filename = source["filename"];
	        this.status = source["status"];
	        this.progress = source["progress"];
	        this.elapsed = source["elapsed"];
	        this.fps = source["fps"];
	        this.originalSize = source["originalSize"];
	        this.outputSize = source["outputSize"];
	        this.savings = source["savings"];
	        this.error = source["error"];
	        this.outputPath = source["outputPath"];
	        this.skipReason = source["skipReason"];
	        this.originalDeleted = source["originalDeleted"];
	        this.addedAt = this.convertValues(source["addedAt"], null);
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

export namespace settings {
	
	export class Settings {
	    encoder: string;
	    crf: number;
	    keepAudio: boolean;
	    discardIfNoGain: boolean;
	    acceptedFormats: string[];
	    outputFolder?: string;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.encoder = source["encoder"];
	        this.crf = source["crf"];
	        this.keepAudio = source["keepAudio"];
	        this.discardIfNoGain = source["discardIfNoGain"];
	        this.acceptedFormats = source["acceptedFormats"];
	        this.outputFolder = source["outputFolder"];
	    }
	}

}

