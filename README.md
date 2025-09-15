<p align="center"><img src="logo.svg" width="40%" border="0" alt="avo"/></p>

# `dispear`

`dispear` is an Elasticsearch ingest pipeline generation library. It makes it easier to write and maintain consistent Elasticsearch ingest pipelines by reducing repetitive code and enabling shared pipeline syntax usage.

- Repeated/themed syntax in an ingest pipeline can be expressed using for loops, for example
    ```
    	for _, s := range []struct{ dst, src string }{
    		{dst: "cloud.account.id", src: "ds.cloud.account.uid"},
    		{dst: "cloud.account.id", src: "ds.cloud.account.uid"},
    		{dst: "cloud.account.name", src: "ds.cloud.account.name"},
    		{dst: "cloud.availability_zone", src: "ds.cloud.zone"},
    		{dst: "cloud.project.id", src: "ds.cloud.project_uid"},
    		{dst: "cloud.provider", src: "ds.cloud.provider"},
    		{dst: "cloud.region", src: "ds.cloud.region"},
    	} {
    		SET(s.dst).COPY_FROM(s.src).IGNORE_EMPTY(true)
    	}
    ```
    expands to
    ```
      - set:
          tag: set_cloud_account_id_1
          field: cloud.account.id
          copy_from: ds.cloud.account.uid
          ignore_empty_value: true
      - set:
          tag: set_cloud_account_id_2
          field: cloud.account.id
          copy_from: ds.cloud.account.uid
          ignore_empty_value: true
      - set:
          tag: set_cloud_account_name
          field: cloud.account.name
          copy_from: ds.cloud.account.name
          ignore_empty_value: true
      - set:
          tag: set_cloud_availability_zone
          field: cloud.availability_zone
          copy_from: ds.cloud.zone
          ignore_empty_value: true
      - set:
          tag: set_cloud_project_id
          field: cloud.project.id
          copy_from: ds.cloud.project_uid
          ignore_empty_value: true
      - set:
          tag: set_cloud_provider
          field: cloud.provider
          copy_from: ds.cloud.provider
          ignore_empty_value: true
      - set:
          tag: set_cloud_region
          field: cloud.region
          copy_from: ds.cloud.region
          ignore_empty_value: true
    ```
- **All** processors are identified with unique processor tags, either user specified or automatic.
- Shared logic can be simply expressed allowing all instances to be updated in unison.
    ```
    // removeErrorHandler is a common error handler that removes a field if an error occurs.
    func removeErrorHandler(f string) []Renderer {
    	return []Renderer{
    		REMOVE(f).IGNORE_MISSING(true),
    		APPEND("error.message", errorFormat),
    	}
    }
    
    const errorFormat = "Processor {{{_ingest.on_failure_processor_type}}} with tag {{{_ingest.on_failure_processor_tag}}} in pipeline {{{_ingest.on_failure_pipeline}}} failed with message: {{{_ingest.on_failure_message}}}"
    ```
    ```
    	LOWERCASE("event.action", "ds.activity_name").
    		IGNORE_MISSING(true).
    		ON_FAILURE(removeErrorHandler("ds.activity_name")...)
    	GSUB("", "event.action", "[: ]", "-").
    		IGNORE_MISSING(true).
    		ON_FAILURE(removeErrorHandler("event.action")...)
    	SET("event.code").
    		COPY_FROM("ds.metadata.event_code").
    		IGNORE_EMPTY(true)
    	CONVERT("", "ds.duration", "long").
    		IGNORE_MISSING(true).
    		ON_FAILURE(removeErrorHandler("ds.duration")...)
    ```


A basic `dispear` program looks like this
```
package main

import (
	. "github.com/efd6/dispear"
)

const (
	ECSVersion = "8.11.0"
	PkgRoot    = "datastream_name"
)

func main() {
	DESCRIPTION("Pipeline for processing Amazon Security Lake Events.")
	ON_FAILURE(
		SET("event.pipeline").VALUE("pipeline_error"),
		APPEND("event.message", errorFormat),
		APPEND("tags", "preserve_original_event").
			ALLOW_DUPLICATES(false),
	)

	SET("ecs.version").VALUE(ECSVersion)
	RENAME("message", "event.original").
		IF("ctx.event?.original == null").
		IGNORE_MISSING(true)
	JSON(PkgRoot, "event.original").ON_FAILURE(
		APPEND("error.message", errorFormat),
	)

	// Many more processors.

	Generate()
}

const errorFormat = "Processor {{{_ingest.on_failure_processor_type}}} with tag {{{_ingest.on_failure_processor_tag}}} in pipeline {{{_ingest.on_failure_pipeline}}} failed with message: {{{_ingest.on_failure_message}}}"
```

For a more complex, complete pipeline, take a look at the [real](./testdata/real.txt) test case which shows the Amazon Security Lake ingest pipeline generated from `dispear`.

Documentation for the processor renderers is available [here](https://pkg.go.dev/github.com/efd6/dispear).
