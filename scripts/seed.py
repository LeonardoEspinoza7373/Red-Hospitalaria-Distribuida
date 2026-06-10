#!/usr/bin/env python3
"""
Seeder para la Red Hospitalaria Distribuida.

Lee archivos JSON con datos de demostración y los envía al backend
a través de la API REST, resolviendo automáticamente las referencias
entre entidades (_ref_id -> ID real).

Uso:
    python3 seed.py                          # localhost:8080 con admin/admin
    python3 seed.py --url http://192.168.1.10:8080
    python3 seed.py --username admin --password admin
    python3 seed.py --url URL --username USER --password PASS
"""

import json
import os
import sys
import time
import urllib.request
import urllib.error

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
DATA_DIR = os.path.join(SCRIPT_DIR, "seed_data")

API_URL = "http://localhost:8080"
API_USERNAME = "admin"
API_PASSWORD = "admin"

ENTITIES = ["usuarios", "pacientes", "donantes", "organos", "trasplantes"]

REFERENCES = {
    "organos": [("donantes", "donante_id")],
    "trasplantes": [("pacientes", "paciente_id"), ("organos", "organo_id")],
}

REQUIRED_FIELDS = {
    "usuarios": ["username", "password", "display_name", "role", "hospital_id"],
    "pacientes": ["nombre", "tipo_sangre", "prioridad", "hospital_id"],
    "donantes": ["nombre", "tipo_sangre", "edad", "detalle", "hospital_id"],
    "organos": ["tipo", "compatibilidad", "estado", "donante_id", "hospital_id"],
    "trasplantes": ["paciente_id", "organo_id", "fecha", "estado", "hospital_id"],
}

id_map = {}


def info(msg):
    print(f"  {msg}")


def ok(msg):
    print(f"  [+] {msg}")


def fail(msg):
    print(f"  [!] {msg}")


def die(msg):
    print(f"[ERROR] {msg}")
    sys.exit(1)


def parse_args():
    global API_URL, API_USERNAME, API_PASSWORD
    i = 1
    while i < len(sys.argv):
        a = sys.argv[i]
        if a == "--url" and i + 1 < len(sys.argv):
            API_URL = sys.argv[i + 1].rstrip("/")
            i += 2
        elif a == "--username" and i + 1 < len(sys.argv):
            API_USERNAME = sys.argv[i + 1]
            i += 2
        elif a == "--password" and i + 1 < len(sys.argv):
            API_PASSWORD = sys.argv[i + 1]
            i += 2
        elif a == "--help":
            print(__doc__)
            sys.exit(0)
        else:
            die(f"Argumento desconocido: {a}")


def login():
    url = f"{API_URL}/api/login"
    body = json.dumps({"username": API_USERNAME, "password": API_PASSWORD}).encode()
    req = urllib.request.Request(url, data=body, headers={"Content-Type": "application/json"})
    try:
        with urllib.request.urlopen(req) as r:
            data = json.loads(r.read())
            ok(f"Conectado como '{data['user']['display_name']}' ({data['user']['role']})")
            return data["token"]
    except urllib.error.HTTPError as e:
        die(f"Error de autenticacion: {e.code} - {e.read().decode()}")


def api(method, path, data=None, token=None):
    url = f"{API_URL}{path}"
    headers = {"Content-Type": "application/json"}
    if token:
        headers["Authorization"] = f"Bearer {token}"
    payload = json.dumps(data, ensure_ascii=False).encode() if data else None
    req = urllib.request.Request(url, data=payload, headers=headers, method=method)
    try:
        with urllib.request.urlopen(req) as r:
            raw = r.read()
            return json.loads(raw) if raw else None
    except urllib.error.HTTPError as e:
        return None


def resolve_references(item, entity):
    if entity not in REFERENCES:
        return
    for src_entity, field in REFERENCES[entity]:
        ref = item.get(field)
        if ref is not None:
            mapping = id_map.get(src_entity, {})
            if ref in mapping:
                item[field] = mapping[ref]
            else:
                fail(f"Referencia {field}={ref} no encontrada en {src_entity}")


def send_entity(token, entity, item):
    ref_id = item.pop("_ref_id", None)
    resolve_references(item, entity)

    if entity == "usuarios":
        path = "/api/usuarios"
    else:
        path = f"/api/{entity}"

    result = api("POST", path, data=item, token=token)
    if result is not None:
        real_id = result.get("id", result.get("ID"))
        if ref_id is not None:
            id_map.setdefault(entity, {})[ref_id] = real_id
        ok(f"{entity[:-1].capitalize()} ID {real_id} creado")
        return True
    else:
        fail(f"Error creando {entity[:-1]} (ref_id={ref_id})")
        return False


def load_json(entity):
    path = os.path.join(DATA_DIR, f"{entity}.json")
    if not os.path.exists(path):
        return None
    with open(path, "r", encoding="utf-8") as f:
        return json.load(f)


def validate_item(item, entity):
    missing = [f for f in REQUIRED_FIELDS.get(entity, []) if f not in item]
    if missing:
        fail(f"Campos requeridos faltantes en {entity}: {missing}")
        return False
    return True


def main():
    parse_args()

    if not os.path.isdir(DATA_DIR):
        die(f"Directorio de datos no encontrado: {DATA_DIR}")

    print(f"\n{'='*60}")
    print(f"  RED HOSPITALARIA DISTRIBUIDA - SEEDER")
    print(f"  API: {API_URL}")
    print(f"{'='*60}\n")

    token = login()

    total_ok = 0
    total_fail = 0

    for entity in ENTITIES:
        items = load_json(entity)
        if items is None:
            info(f"'{entity}.json' no encontrado, se omite.")
            continue
        if not items:
            info(f"'{entity}' sin datos, se omite.")
            continue

        print(f"\n--- Creando {len(items)} {entity} ---")
        for item in items:
            if not validate_item(item, entity):
                total_fail += 1
                continue
            if send_entity(token, entity, item):
                total_ok += 1
            else:
                total_fail += 1
            time.sleep(0.15)

    print(f"\n{'='*60}")
    print(f"  RESUMEN")
    print(f"  Creados:   {total_ok}")
    print(f"  Fallidos:  {total_fail}")
    print(f"{'='*60}")

    if id_map:
        print(f"\n  Mapeo de referencias (_ref_id -> ID real):")
        for entity, mapping in id_map.items():
            if mapping:
                print(f"    {entity}:")
                for ref, real in sorted(mapping.items()):
                    print(f"      _ref_id {ref:>3} -> ID real {real}")

    print()


if __name__ == "__main__":
    main()
