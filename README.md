# workflow-advisor
Contains CLI and workflow which scans user repositories to detect technologies, and generates starter workflow

### Usage

Single tech detection:

```bash
 advisor -g js -w "${PWD}/test-workflow.yaml" --src "${PWD}"
```

Multiple tech detection:

```bash
 advisor -g js -g csharp -w "${PWD}/test-workflow.yaml" --src "${PWD}"
```

### Output

Output is defined by json scheme [here](advisor-output.scheme.json)


### License

This code is made available under the 
[MIT license](https://opensource.org/license/mit/.)

### References 

* Learn about [the CloudBees platform](https://docs.cloudbees.com/docs/cloudbees-saas-platform/latest/.)