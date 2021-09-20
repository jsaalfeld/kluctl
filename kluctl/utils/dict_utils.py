import fnmatch
import threading

from jsonpath_ng import auto_id_field, jsonpath, JSONPath
from jsonpath_ng.ext import parse

json_path_cache = {}
json_path_cache_mutex = threading.Lock()

# Add better wildcard support to jsonpath lib
def ext_reified_fields(self, datum):
    result = []
    for field in self.fields:
        if "*" not in field:
            result.append(field)
            continue
        try:
            fields = [f for f in datum.value.keys() if fnmatch.fnmatch(f, field)]
            if auto_id_field is not None:
                fields.append(auto_id_field)
            result += fields
        except AttributeError:
            pass
    return tuple(result)
jsonpath.Fields.reified_fields = ext_reified_fields

def parse_json_path(p) -> JSONPath:
    if isinstance(p, list):
        p2 = "$"
        for x in p:
            if isinstance(x, str):
                if '"' in x:
                    p2 = "%s['%s']" % (p2, x)
                else:
                    p2 = '%s["%s"]' % (p2, x)
            else:
                p2 = "%s[%d]" % (p2, x)
        p = p2

    with json_path_cache_mutex:
        if p in json_path_cache:
            return json_path_cache[p]
        pp = parse(p)
        json_path_cache[p] = pp
        return pp

def copy_primitive_value(v):
    if isinstance(v, dict):
        return copy_dict(v)
    if isinstance(v, str) or isinstance(v, bytes):
        return v
    if is_iterable(v):
        return [copy_primitive_value(x) for x in v]
    return v

def copy_dict(a):
    ret = {}
    for k, v in a.items():
        ret[k] = copy_primitive_value(v)
    return ret

def merge_dict(a, b, clone=True):
    if clone:
        a = copy_dict(a)
    if a is None:
        a = {}
    if b is None:
        b = {}
    for key in b:
        if key in a:
            if isinstance(a[key], dict) and isinstance(b[key], dict):
                merge_dict(a[key], b[key], clone=False)
            else:
                a[key] = b[key]
        else:
            a[key] = b[key]
    return a

def set_dict_default_value(d, path, default):
    p = parse_json_path(path)
    f = p.find(d)

    if len(f) > 1:
        raise Exception("Only simple jsonpath supported in set_dict_default_value")
    if len(f) == 0 or f[0].value is None:
        p.update_or_create(d, default)

def get_dict_value(y, path, default=None):
    p = parse_json_path(path)
    f = p.find(y)
    if len(f) > 1:
        raise Exception("Only simple jsonpath supported in get_dict_value")
    if len(f) == 0:
        return default
    return f[0].value

def set_dict_value(y, path, value, do_clone=False):
    if do_clone:
        y = copy_dict(y)
    p = parse_json_path(path)
    p.update_or_create(y, value)
    return y

def del_dict_value(d, k):
    p = parse_json_path(k)
    p.filter(lambda x: True, d)

def is_empty(o):
    if isinstance(o, dict) or isinstance(o, list):
        return len(o) == 0
    return False

def remove_empty(o):
    if isinstance(o, dict):
        for k in list(o.keys()):
            remove_empty(o[k])
            if is_empty(o[k]):
                del(o[k])
    elif isinstance(o, list):
        i = 0
        while i < len(o):
            v = o[i]
            remove_empty(v)
            if is_empty(v):
                o.pop(i)
            else:
                i += 1

def is_iterable(obj):
    if isinstance(obj, list):
        return True
    if isinstance(obj, tuple):
        return True
    if isinstance(obj, dict):
        return True
    if isinstance(obj, str):
        return True
    if isinstance(obj, bytes):
        return True
    if isinstance(obj, int) or isinstance(obj, bool):
        return False
    if isinstance(obj, type):
        return False
    try:
        iter(obj)
    except Exception:
        return False
    else:
        return True

def object_iterator(o):
    stack = [(o, [])]

    while len(stack) != 0:
        o2, p = stack.pop()
        yield o2, p

        if isinstance(o2, dict):
            for k, v in o2.items():
                stack.append((v, p + [k]))
        elif not isinstance(o2, str) and is_iterable(o2):
            for i, v in enumerate(o2):
                stack.append((v, p + [i]))
