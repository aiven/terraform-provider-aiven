import json
import sys


class DocGen:

    KEYS = ['provider', 'resource_schemas', 'datasource_schemas']

    def __init__(self, schema = "scripts/schema.json", output = "ALLL"):
        with open(schema) as source:
            json_src = json.load(source)
            md_str = []
            if output is "ALL":
                md_str.append("# terraform-provider-aiven \n")
                c, r, o = self.get_attrs(json_src['provider'])
                md_str += self.gen_attributes(c, r, o)
                md_str.append('--- \n')
            if output in ["ALL", "RESOURCES"]:
                md_str.append('## Resources \n')
                for k, v in json_src['resource_schemas'].items():
                    c, r, o = self.get_attrs(v)
                    md_str += self.gen_attributes(c, r, o, k)
                md_str.append('--- \n')
            if output in ["ALL", "DATASOURCES"]:
                md_str.append('## Data-sources \n')
                for k, v in json_src['data_source_schemas'].items():
                    c, r, o = self.get_attrs(v)
                    md_str += self.gen_attributes(c, r, o, k)

            print(''.join(md_str))

    def get_attrs(self, obj):
        c = []
        r = []
        o = []
        attrs = obj['block']['attributes']
        for k, v in attrs.items():
            desc = "_{}_".format(v['description']) if 'description' in v else ''
            s = "**{}** {}".format(k, desc)
            if 'computed' in v and v['computed']:
                c.append(s)
            elif 'required' in v and v['required']:
                r.append(s)
            elif 'optional' in v and v['optional']:
                o.append(s)
        return c, r, o


    def gen_list(self, arr, section=False, numbered=False):
        out = []
        if len(arr) == 0:
            return out
        if section is not False:
            out += section
        for req in arr:
            idx = '*'
            out.append("{} {} \n".format(idx, req))
        return out


    def gen_attributes(self, c, r, o, section=False):
        tmpl = []
        if section is not False:
            tmpl.append("### {} \n".format(section))
        tmpl += self.gen_list(r, '#### Required \n')
        tmpl += self.gen_list(o, '#### Optional \n')
        tmpl += self.gen_list(c, '##### Computed \n')
        return tmpl


if __name__ == '__main__':
    if len(sys.argv) == 3:
        schema = sys.argv[1]
        output = sys.argv[2]
        d = DocGen(schema, output)
    else:
        print("Usage: docgen.py <PATH_TO_SCHEMA> <ALL | RESOURCES | DATASOURCES>")