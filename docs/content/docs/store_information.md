---
title: "Workspace Directory & Store Mapping"
weight: 40
---

# Google Workspace Directory & Store Footprint Mapping Reference

This reference manual documents the internal sandbox Google Workspace directory, Organizational Unit (OU) topologies, store-level personnel matrix, role configurations, and regional footprints used within the **Enterprise Task Engine**.

---

## 1. Google Workspace Directory OU Structure

To test enterprise role-based access controls (RBAC), multi-store operational overlaps, and location-based document grounding (pgvector localized SOP filters), the local sandbox environment implements a fully structured Google Workspace tree. 

Parent and child OUs are organized operationally:

```mermaid
graph TD
    Root["retail.altostrat.com<br/>(Root Domain)"]
    Stores["Stores<br/>(Parent OU)"]
    Regional["Regional Managers<br/>(Regional Directors)"]
    Store1["Volt & Vine - Seattle<br/>(Store Footprint OU)"]
    Store2["Volt & Vine - San Francisco<br/>(Store Footprint OU)"]
    Store3["OmniMart - Store #1000<br/>(Dallas Store Footprint OU)"]

    SubRoles1["Admins | Managers | Cashiers | Associates | Vault"]
    SubRoles2["Admins | Managers | Cashiers | Associates | Vault"]
    SubRoles3["Admins | Managers | Cashiers | Associates | Vault"]

    Root --> Stores
    Stores --> Regional
    Stores --> Store1
    Stores --> Store2
    Stores --> Store3

    Store1 --> SubRoles1
    Store2 --> SubRoles2
    Store3 --> SubRoles3
```

---

## 2. Active Test Personnel Identity Matrix

The sandbox Workspace directory has been provisioned to support **553 role profiles** spanning **109 physical storefront locations** and **4 regional managers**.

### User Account Naming Conventions
* **Store Level Users:** `{role_slug}-{store_slug}@retail.altostrat.com`
* **Regional Managers:** `regional-manager-{region_slug}@retail.altostrat.com`
* **Initial Passwords Registry:** All temporary credentials and initial login passwords are cataloged inside the secure, git-ignored file `passwords_registry.csv`.

### Active Regional Directors Mapping
Due to sandbox limitations, the **6 retail regions** are mapped to the **4 active regional profiles** inside the core Workspace directory to prevent authentication constraints:

| Retail Footprint Region | Active Workspace Manager Account | Scopes & Responsibilities |
| :--- | :--- | :--- |
| `northeast` | `regional-manager-northeast@retail.altostrat.com` | Northeast Footprint |
| `southeast` | `regional-manager-southeast@retail.altostrat.com` | Southeast Footprint |
| `northwest` | `regional-manager-west@retail.altostrat.com` | Pacific Northwest Footprint |
| `southwest` | `regional-manager-west@retail.altostrat.com` | Pacific Southwest Footprint |
| `northcentral` | `regional-manager-midwest@retail.altostrat.com` | Northcentral Footprint |
| `southcentral` | `regional-manager-midwest@retail.altostrat.com` | Southcentral Footprint |

---

## 3. Complete Store & Region Mapping Table

The complete lookup matrix below identifies all 109 physical store entities, their assigned regional directors, slugs, and unique ID prefixes:

| Region | Store ID Prefix | Physical Store Name | Store Slug | Mapped Regional Director |
| :--- | :--- | :--- | :--- | :--- |
| `northcentral` | `55555555...0009` | OmniMart - Store #1009 (Columbus) | `omnimart-store-1009-columbus` | `regional-manager-midwest` |
| `northcentral` | `55555555...000a` | OmniMart - Store #1010 (Indianapolis) | `omnimart-store-1010-indianapolis` | `regional-manager-midwest` |
| `northcentral` | `55555555...0013` | OmniMart - Store #1019 (Detroit) | `omnimart-store-1019-detroit` | `regional-manager-midwest` |
| `northcentral` | `55555555...001a` | OmniMart - Store #1026 (Milwaukee) | `omnimart-store-1026-milwaukee` | `regional-manager-midwest` |
| `northcentral` | `55555555...0029` | OmniMart - Store #1041 (Minneapolis) | `omnimart-store-1041-minneapolis` | `regional-manager-midwest` |
| `northcentral` | `55555555...0030` | OmniMart - Store #1048 (Cleveland) | `omnimart-store-1048-cleveland` | `regional-manager-midwest` |
| `northcentral` | `55555555...0031` | OmniMart - Store #1049 (Aurora) | `omnimart-store-1049-aurora` | `regional-manager-midwest` |
| `northcentral` | `55555555...0039` | OmniMart - Store #1057 (Stockton) | `omnimart-store-1057-stockton` | `regional-manager-midwest` |
| `northcentral` | `55555555...003a` | OmniMart - Store #1058 (Saint Paul) | `omnimart-store-1058-saint-paul` | `regional-manager-midwest` |
| `northcentral` | `55555555...003b` | OmniMart - Store #1059 (Cincinnati) | `omnimart-store-1059-cincinnati` | `regional-manager-midwest` |
| `northcentral` | `55555555...003f` | OmniMart - Store #1063 (Lincoln) | `omnimart-store-1063-lincoln` | `regional-manager-midwest` |
| `northcentral` | `55555555...0047` | OmniMart - Store #1071 (Toledo) | `omnimart-store-1071-toledo` | `regional-manager-midwest` |
| `northcentral` | `55555555...0048` | OmniMart - Store #1072 (Fort Wayne) | `omnimart-store-1072-fort-wayne` | `regional-manager-midwest` |
| `northcentral` | `55555555...004d` | OmniMart - Store #1077 (Madison) | `omnimart-store-1077-madison` | `regional-manager-midwest` |
| `northcentral` | `55555555...0060` | OmniMart - Store #1096 (Des Moines) | `omnimart-store-1096-des-moines` | `regional-manager-midwest` |
| `northcentral` | `44444444...0005` | Volt & Vine - Chicago | `volt-and-vine-chicago` | `regional-manager-midwest` |
| `northeast` | `55555555...0004` | OmniMart - Store #1004 (Philadelphia) | `omnimart-store-1004-philadelphia` | `regional-manager-northeast` |
| `northeast` | `55555555...000f` | OmniMart - Store #1015 (Washington) | `omnimart-store-1015-washington` | `regional-manager-northeast` |
| `northeast` | `55555555...0010` | OmniMart - Store #1016 (Boston) | `omnimart-store-1016-boston` | `regional-manager-northeast` |
| `northeast` | `55555555...0019` | OmniMart - Store #1025 (Baltimore) | `omnimart-store-1025-baltimore` | `regional-manager-northeast` |
| `northeast` | `55555555...0044` | OmniMart - Store #1068 (Newark) | `omnimart-store-1068-newark` | `regional-manager-northeast` |
| `northeast` | `55555555...004b` | OmniMart - Store #1075 (Jersey City) | `omnimart-store-1075-jersey-city` | `regional-manager-northeast` |
| `northeast` | `55555555...0051` | OmniMart - Store #1081 (Buffalo) | `omnimart-store-1081-buffalo` | `regional-manager-northeast` |
| `northeast` | `44444444...0008` | Volt & Vine - Boston | `volt-and-vine-boston` | `regional-manager-northeast` |
| `northeast` | `44444444...0007` | Volt & Vine - New York | `volt-and-vine-new-york` | `regional-manager-northeast` |
| `northwest` | `55555555...000d` | OmniMart - Store #1013 (Seattle) | `omnimart-store-1013-seattle` | `regional-manager-west` |
| `northwest` | `55555555...0015` | OmniMart - Store #1021 (Portland) | `omnimart-store-1021-portland` | `regional-manager-west` |
| `northwest` | `55555555...0040` | OmniMart - Store #1064 (Anchorage) | `omnimart-store-1064-anchorage` | `regional-manager-west` |
| `northwest` | `55555555...005d` | OmniMart - Store #1093 (Boise) | `omnimart-store-1093-boise` | `regional-manager-west` |
| `northwest` | `55555555...005f` | OmniMart - Store #1095 (Spokane) | `omnimart-store-1095-spokane` | `regional-manager-west` |
| `northwest` | `55555555...0061` | OmniMart - Store #1097 (Tacoma) | `omnimart-store-1097-tacoma` | `regional-manager-west` |
| `northwest` | `44444444...0000` | Volt & Vine - Seattle | `volt-and-vine-seattle` | `regional-manager-west` |
| `southcentral` | `55555555...0000` | OmniMart - Store #1000 (Dallas) | `omnimart-store-1000-dallas` | `regional-manager-midwest` |
| `southcentral` | `55555555...0001` | OmniMart - Store #1001 (Houston) | `omnimart-store-1001-houston` | `regional-manager-midwest` |
| `southcentral` | `55555555...0002` | OmniMart - Store #1002 (San Antonio) | `omnimart-store-1002-san-antonio` | `regional-manager-midwest` |
| `southcentral` | `55555555...0007` | OmniMart - Store #1007 (Austin) | `omnimart-store-1007-austin` | `regional-manager-midwest` |
| `southcentral` | `55555555...000e` | OmniMart - Store #1014 (Denver) | `omnimart-store-1014-denver` | `regional-manager-midwest` |
| `southcentral` | `55555555...0011` | OmniMart - Store #1017 (El Paso) | `omnimart-store-1017-el-paso` | `regional-manager-midwest` |
| `southcentral` | `55555555...0014` | OmniMart - Store #1020 (Oklahoma City) | `omnimart-store-1020-oklahoma-city` | `regional-manager-midwest` |
| `southcentral` | `55555555...0018` | OmniMart - Store #1024 (Louisville) | `omnimart-store-1024-louisville` | `regional-manager-midwest` |
| `southcentral` | `55555555...0020` | OmniMart - Store #1032 (Kansas City) | `omnimart-store-1032-kansas-city` | `regional-manager-midwest` |
| `southcentral` | `55555555...0022` | OmniMart - Store #1034 (Omaha) | `omnimart-store-1034-omaha` | `regional-manager-midwest` |
| `southcentral` | `55555555...0023` | OmniMart - Store #1035 (Colorado Springs) | `omnimart-store-1035-colorado-springs` | `regional-manager-midwest` |
| `southcentral` | `55555555...002a` | OmniMart - Store #1042 (Tulsa) | `omnimart-store-1042-tulsa` | `regional-manager-midwest` |
| `southcentral` | `55555555...002e` | OmniMart - Store #1046 (Wichita) | `omnimart-store-1046-wichita` | `regional-manager-midwest` |
| `southcentral` | `55555555...0033` | OmniMart - Store #1051 (Honolulu) | `omnimart-store-1051-honolulu` | `regional-manager-midwest` |
| `southcentral` | `55555555...0036` | OmniMart - Store #1054 (Corpus Christi) | `omnimart-store-1054-corpus-christi` | `regional-manager-midwest` |
| `southcentral` | `55555555...0037` | OmniMart - Store #1055 (Lexington) | `omnimart-store-1055-lexington` | `regional-manager-midwest` |
| `southcentral` | `55555555...003c` | OmniMart - Store #1060 (St. Louis) | `omnimart-store-1060-st-louis` | `regional-manager-midwest` |
| `southcentral` | `55555555...003d` | OmniMart - Store #1061 (Pittsburgh) | `omnimart-store-1061-pittsburgh` | `regional-manager-midwest` |
| `southcentral` | `55555555...0041` | OmniMart - Store #1065 (Plano) | `omnimart-store-1065-plano` | `regional-manager-midwest` |
| `southcentral` | `55555555...004a` | OmniMart - Store #1074 (Laredo) | `omnimart-store-1074-laredo` | `regional-manager-midwest` |
| `southcentral` | `55555555...004e` | OmniMart - Store #1078 (Lubbock) | `omnimart-store-1078-lubbock` | `regional-manager-midwest` |
| `southcentral` | `55555555...0055` | OmniMart - Store #1085 (Winston-Salem) | `omnimart-store-1085-winston-salem` | `regional-manager-midwest` |
| `southcentral` | `55555555...0059` | OmniMart - Store #1089 (Garland) | `omnimart-store-1089-garland` | `regional-manager-midwest` |
| `southcentral` | `55555555...005a` | OmniMart - Store #1090 (Irving) | `omnimart-store-1090-irving` | `regional-manager-midwest` |
| `southcentral` | `55555555...005c` | OmniMart - Store #1092 (Arvada) | `omnimart-store-1092-arvada` | `regional-manager-midwest` |
| `southcentral` | `55555555...0062` | OmniMart - Store #1098 (San Bernardino) | `omnimart-store-1098-san-bernardino` | `regional-manager-midwest` |
| `southcentral` | `44444444...0004` | Volt & Vine - Austin | `volt-and-vine-austin` | `regional-manager-midwest` |
| `southcentral` | `44444444...0003` | Volt & Vine - Denver | `volt-and-vine-denver` | `regional-manager-midwest` |
| `southeast` | `55555555...0008` | OmniMart - Store #1008 (Jacksonville) | `omnimart-store-1008-jacksonville` | `regional-manager-southeast` |
| `southeast` | `55555555...000b` | OmniMart - Store #1011 (Charlotte) | `omnimart-store-1011-charlotte` | `regional-manager-southeast` |
| `southeast` | `55555555...0012` | OmniMart - Store #1018 (Nashville) | `omnimart-store-1018-nashville` | `regional-manager-southeast` |
| `southeast` | `55555555...0017` | OmniMart - Store #1023 (Memphis) | `omnimart-store-1023-memphis` | `regional-manager-southeast` |
| `southeast` | `55555555...0021` | OmniMart - Store #1033 (Atlanta) | `omnimart-store-1033-atlanta` | `regional-manager-southeast` |
| `southeast` | `55555555...0024` | OmniMart - Store #1036 (Raleigh) | `omnimart-store-1036-raleigh` | `regional-manager-southeast` |
| `southeast` | `55555555...0026` | OmniMart - Store #1038 (Virginia Beach) | `omnimart-store-1038-virginia-beach` | `regional-manager-southeast` |
| `southeast` | `55555555...0027` | OmniMart - Store #1039 (Miami) | `omnimart-store-1039-miami` | `regional-manager-southeast` |
| `southeast` | `55555555...002b` | OmniMart - Store #1043 (Tampa) | `omnimart-store-1043-tampa` | `regional-manager-southeast` |
| `southeast` | `55555555...002c` | OmniMart - Store #1044 (Arlington) | `omnimart-store-1044-arlington` | `regional-manager-southeast` |
| `southeast` | `55555555...002d` | OmniMart - Store #1045 (New Orleans) | `omnimart-store-1045-new-orleans` | `regional-manager-southeast` |
| `southeast` | `55555555...003e` | OmniMart - Store #1062 (Greensboro) | `omnimart-store-1062-greensboro` | `regional-manager-southeast` |
| `southeast` | `55555555...0042` | OmniMart - Store #1066 (Orlando) | `omnimart-store-1066-orlando` | `regional-manager-southeast` |
| `southeast` | `55555555...0045` | OmniMart - Store #1069 (Durham) | `omnimart-store-1069-durham` | `regional-manager-southeast` |
| `southeast` | `55555555...0049` | OmniMart - Store #1073 (St. Petersburg) | `omnimart-store-1073-st-petersburg` | `regional-manager-southeast` |
| `southeast` | `55555555...0056` | OmniMart - Store #1086 (Chesapeake) | `omnimart-store-1086-chesapeake` | `regional-manager-southeast` |
| `southeast` | `55555555...0057` | OmniMart - Store #1087 (Norfolk) | `omnimart-store-1087-norfolk` | `regional-manager-southeast` |
| `southeast` | `55555555...005b` | OmniMart - Store #1091 (Hialeah) | `omnimart-store-1091-hialeah` | `regional-manager-southeast` |
| `southeast` | `55555555...005e` | OmniMart - Store #1094 (Richmond) | `omnimart-store-1094-richmond` | `regional-manager-southeast` |
| `southeast` | `44444444...0006` | Volt & Vine - Atlanta | `volt-and-vine-atlanta` | `regional-manager-southeast` |
| `southeast` | `44444444...0009` | Volt & Vine - Miami | `volt-and-vine-miami` | `regional-manager-southeast` |
| `southwest` | `55555555...0003` | OmniMart - Store #1003 (Phoenix) | `omnimart-store-1003-phoenix` | `regional-manager-west` |
| `southwest` | `55555555...0005` | OmniMart - Store #1005 (San Diego) | `omnimart-store-1005-san-diego` | `regional-manager-west` |
| `southwest` | `55555555...0006` | OmniMart - Store #1006 (San Jose) | `omnimart-store-1006-san-jose` | `regional-manager-west` |
| `southwest` | `55555555...000c` | OmniMart - Store #1012 (San Francisco) | `omnimart-store-1012-san-francisco` | `regional-manager-west` |
| `southwest` | `55555555...0016` | OmniMart - Store #1022 (Las Vegas) | `omnimart-store-1022-las-vegas` | `regional-manager-west` |
| `southwest` | `55555555...001b` | OmniMart - Store #1027 (Albuquerque) | `omnimart-store-1027-albuquerque` | `regional-manager-west` |
| `southwest` | `55555555...001c` | OmniMart - Store #1028 (Tucson) | `omnimart-store-1028-tucson` | `regional-manager-west` |
| `southwest` | `55555555...001d` | OmniMart - Store #1029 (Fresno) | `omnimart-store-1029-fresno` | `regional-manager-west` |
| `southwest` | `55555555...001e` | OmniMart - Store #1030 (Sacramento) | `omnimart-store-1030-sacramento` | `regional-manager-west` |
| `southwest` | `55555555...001f` | OmniMart - Store #1031 (Mesa) | `omnimart-store-1031-mesa` | `regional-manager-west` |
| `southwest` | `55555555...0025` | OmniMart - Store #1037 (Long Beach) | `omnimart-store-1037-long-beach` | `regional-manager-west` |
| `southwest` | `55555555...0028` | OmniMart - Store #1040 (Oakland) | `omnimart-store-1040-oakland` | `regional-manager-west` |
| `southwest` | `55555555...002f` | OmniMart - Store #1047 (Bakersfield) | `omnimart-store-1047-bakersfield` | `regional-manager-west` |
| `southwest` | `55555555...0032` | OmniMart - Store #1050 (Anaheim) | `omnimart-store-1050-anaheim` | `regional-manager-west` |
| `southwest` | `55555555...0034` | OmniMart - Store #1052 (Santa Ana) | `omnimart-store-1052-santa-ana` | `regional-manager-west` |
| `southwest` | `55555555...0035` | OmniMart - Store #1053 (Riverside) | `omnimart-store-1053-riverside` | `regional-manager-west` |
| `southwest` | `55555555...0038` | OmniMart - Store #1056 (Henderson) | `omnimart-store-1056-henderson` | `regional-manager-west` |
| `southwest` | `55555555...0043` | OmniMart - Store #1067 (Irvine) | `omnimart-store-1067-irvine` | `regional-manager-west` |
| `southwest` | `55555555...0046` | OmniMart - Store #1070 (Chula Vista) | `omnimart-store-1070-chula-vista` | `regional-manager-west` |
| `southwest` | `55555555...004c` | OmniMart - Store #1076 (Chandler) | `omnimart-store-1076-chandler` | `regional-manager-west` |
| `southwest` | `55555555...004f` | OmniMart - Store #1079 (Scottsdale) | `omnimart-store-1079-scottsdale` | `regional-manager-west` |
| `southwest` | `55555555...0050` | OmniMart - Store #1080 (Reno) | `omnimart-store-1080-reno` | `regional-manager-west` |
| `southwest` | `55555555...0052` | OmniMart - Store #1082 (Gilbert) | `omnimart-store-1082-gilbert` | `regional-manager-west` |
| `southwest` | `55555555...0053` | OmniMart - Store #1083 (Glendale) | `omnimart-store-1083-glendale` | `regional-manager-west` |
| `southwest` | `55555555...0054` | OmniMart - Store #1084 (North Las Vegas) | `omnimart-store-1084-north-las-vegas` | `regional-manager-west` |
| `southwest` | `55555555...0058` | OmniMart - Store #1088 (Fremont) | `omnimart-store-1088-fremont` | `regional-manager-west` |
| `southwest` | `55555555...0063` | OmniMart - Store #1099 (Modesto) | `omnimart-store-1099-modesto` | `regional-manager-west` |
| `southwest` | `44444444...0002` | Volt & Vine - Los Angeles | `volt-and-vine-los-angeles` | `regional-manager-west` |
| `southwest` | `44444444...0001` | Volt & Vine - San Francisco | `volt-and-vine-san-francisco` | `regional-manager-west` |

---

## 4. Directory Sub-OU Role Access Mapping
Inside each retail footprint OU, directory accounts are distributed under five standard sub-OUs mapped to private Relational Database GTE Role IDs:

| Directory Sub-OU | Role Slug | Typical Operational Job Title | internal GTE Database Role ID |
| :--- | :--- | :--- | :--- |
| `Admins` | `admin` | Store Systems Administrator | `ROLE_SITE_MANAGER` |
| `Managers` | `manager` | Store Operations Manager | `ROLE_SITE_MANAGER` |
| `Cashiers` | `cashier` | Store Cashier | `ROLE_SITE_ASSOCIATE` |
| `Associates` | `associate` | Customer Support Associate | `ROLE_SITE_ASSOCIATE` |
| `Vault` | `vault` | Vault Cash Custodian | `ROLE_SITE_ASSOCIATE` |
