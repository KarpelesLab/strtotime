#!/bin/bash
# Regenerate testdata/date_parse_php.jsonl from the input list.
# Each line is {"in": "<input>", "out": <php_date_parse_output>}.
#
# Usage: ./testdata/date_parse_php.sh > testdata/date_parse_php.jsonl
set -e

cat <<'EOF' | php -r '
$stdin = stream_get_contents(STDIN);
foreach (explode("\n", trim($stdin)) as $c) {
    if ($c === "") continue;
    echo json_encode(["in"=>$c,"out"=>date_parse($c)], JSON_UNESCAPED_SLASHES) . "\n";
}'
2006-12-12T10:00:00.5+01:00
2023-01-15
2023-01-15T14:30:45Z
2023-01-15T14:30:45+09:00
2023-01-15T14:30:45-05:00
2023-01-01 Asia/Tokyo
2023-01-01 America/New_York
2023-01-01 EST
2023-01-01 GMT
2023-01-01 UTC
2023-01-01 PDT
May 2020
Oct 2001
January 15 2023
Jan 15, 2023
26 Nov 2005
26th Nov 2005
2023-W03
2023-W03-1
10:00:00
14:30
10:00:00 AM
2:30 PM
1999/12/31
12/31/1999
31.12.1999
20230115
20230115143045
@1121373041
@1121373041.5
2023-05-30 -1 month
2022-01-01 +1 year
+3 days
-1 week
3 weeks
+1 month
next monday
last friday
this tuesday
next week
last year
first day of next month
last day of this year
first day of January 2023
last day of December 2023
Monday
tuesday
2 weeks ago
3 mondays ago
midnight
noon
tomorrow
yesterday
today
now
2023-01-15 -1 month
2023-05-30 +2 months
2004-10-31 EDT +3 hours
Thu, 31 Jul 2025 11:16:38 +0900
Fri, 07 Dec 2007 19:05:14 +1000
Mon, 08 May 2006 13:06:44 -0400
Sun 2017-01-01
26 Nov 2005 00:00 UTC
June 1 1985 16:30:00 Europe/Paris
2023-01-01 10:30
2023-01-01 10:30:45
2023-01-01 10:30:45.123
01:02:03.456
2023/01/15
01/15/2023
15.01.2023
20230115
+1 hour
-30 minutes
-45 seconds
+1 week
2 hours
3 months
2 years
December 2008
Dec 25
5 minutes ago
1 day ago
2 weeks
8am
11pm
next year
last month
this week
EOF
