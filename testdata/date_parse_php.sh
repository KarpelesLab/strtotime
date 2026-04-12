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
2024-02-29
2023-12-31T23:59:59
2023-12-31T23:59:59Z
2023-12-31T23:59:59+00:00
2023-12-31T23:59:59-10:30
1970-01-01T00:00:00Z
1999-12-31 23:59:59
2023-01-15T14:30:45.999999999
2023-01-15T14:30:45.000001
@0
@1000000000
@-1
Monday noon
January 1st
1st January
February 29 2024
10/10/2010
2010-10-10
10.10.2010
Dec 25 2023
1 year ago
5 years
14:30:00Z
2023-01-01T00:00:00
2024-02-29T12:00:00
2023-02-28T23:59:59
20230101
20230101T123456
10:00 UTC
14:30 PST
14:30 +09:00
14:30:00 +09:00
14:30:00 -05:00
+0900
-0500
2020-01-01 12:00:00 JST
Jul 4 2020 10:30 AM
July 4 2020 10:30 PM
July 4, 2020
2023-01-01 00:00 Z
2023-01-01 00:00:00.5 Z
22-11-2023
Jan 5
Feb 28
March 1st
June 30th 2023
last week
this year
+1 minute
-1 second
2 hours ago
30 minutes ago
20200101 10:00
2023-07-15 14:30:45 America/Los_Angeles
2023-07-15T14:30:45 Asia/Tokyo
2023-07-15T14:30:45.123456
2000-01-01T00:00:00
2038-01-19T03:14:07
1901-12-13T20:45:52
2023-06-15 14:00
Jan 1, 1970
December 31, 2099
15th June 2023
23rd Feb 1999
Feb 23, 1999
Saturday, November 25 2023
November 2023
2023W01
2023W011
+2 weeks 3 days
1 day 2 hours
5 hours 30 minutes
3 hours 15 minutes 5 seconds
+3 days 10:00:00
2023-01-01T10:00:00.123456Z
2023-01-01T10:00:00.123456789Z
2023-01-01T10:00:00.1Z
2023-01-01T10:00:00.12+09:00
1.1.2020
1.1.2020 12:00
31-12-1999
01-Jan-2023
1-Jan-23
2023-1-1
2023-01-1
2023-1-01
1-1-2023
1/1/2023
12/1/2023
00:00:00
00:00
23:59:59
12:00 PM
12:00 AM
12:30 AM
12:30 PM
1:00 am
11:59 pm
00:00:00 UTC
23:59:59 +0100
+00:00
-12:00
Z
UTC
GMT
EST
PST
Asia/Tokyo
Europe/London
America/New_York
first Tuesday of December 2020
second Monday of March 2023
last Friday of July 2024
third thursday
2020-001
2020-W01
2020-W53
2020-W01-7
2020-W53-1
2020-02-29T12:00:00Z
EOF
