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
2023-02-29
2023-00-15
2023-04-31
Feb 29 2023
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
24:00:00
24:00
tomorrow noon
noon tomorrow
yesterday noon
tomorrow midnight
10am tomorrow
tomorrow 10am
monday 10am
noon Monday
10:00 tomorrow
Monday noon
YYYY
foo
nope
2023
1234
1200
0000
0030
2359
2400
9999
Monday 2023-01-01
2020-366
2020-001
2024-366
2023-365
2020-060
2020-100
2023-01-01 +3 days 10:00:00
2023-01-01 10:00:00 +3 days
-1 week +2 days
+1 year +3 months
+2 weeks -1 day
-3 hours +10 minutes
+1 day -2 hours
-1 month +15 days
@1234567890 Asia/Tokyo
@1234567890 UTC
@1000 +0900
23
123
12345
0
99
Friday tomorrow
yesterday Monday
tomorrow noon EST
noon UTC
midnight GMT
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
last Monday of December 2008
last Sunday of August 2020
fourth Wednesday of October 2025
fifth Monday of March 2024
last Saturday of February 2024
01:02:03
11:22
5:05:05
5am
5pm
8:30am
8:30 am
8:30 a.m.
8:30 p.m.
12:30:45.678
12:30:45.678901
22:49:12.42 GMT
22:49:12.42+0200
22:49:12.42+02:00
-00:05:00
+05:45
Asia/Kolkata
Pacific/Honolulu
2023-03-12T02:30:00 America/New_York
2023-11-05T01:30:00 America/New_York
2023-01-15 14:30:45.5
Tue, 15 Feb 2022 18:00:00 GMT
Mon, 2 Jan 2023 09:15:00 JST
Sun, 31 Dec 2023 23:59:59 +0000
1970
9999
@-1000000000
2023-01-15T10:30:45.5+09:00
2023-01-15T10:30:45.5-05:00
2023-01-15T10:30:45+0900
2023-01-15T10:30:45-0500
2023-01-15T10:30:45+09
2023-01-15T10:30:45-05
2023-01-15T10:30+09:00
2023-01-15T10:30Z
2023-01-15T10Z
2023-01-15T10
01-02-2003
1-2-3
15 Jan 2023 10:30
15 Jan 2023 10:30:45
15-Jan-2023
15-Jan
Jan-15
24-Jan-2019
1 January 2023
next day
last day
this minute
this second
+0 seconds
-0 minutes
6 months ago
1 hour ago
36 hours
2020-1-1
2023-02-28T23:59:59
2023-02-28T23:59:59.999
01-Jan-1970
01 Jan 1970
1970-01
2020-12
2000-01-01
2000-02-29
2099-12-31
00:00:00Z
01:00:00 UTC
12:34:56 Asia/Tokyo
23:59:59-12:00
20060502T230245-05
20060502T230245+0200
20060502T230245Z
2006-05-02T23:02:45
2006-05-02T23:02:45Z
2006-05-02T23:02:45+00:00
2038-01-19T03:14:08
-0100
-05:30
+12:00
+14:00
-12:00
Africa/Cairo
Australia/Sydney
Pacific/Auckland
America/Argentina/Buenos_Aires
Indian/Maldives
Atlantic/Azores
Jan 2020
Feb 2020
Apr 2020
Sep 2020
Dec 2020
1st Jan 2020
2nd feb 2020
3rd mar 2020
21st June 2023
last Monday
last Sunday
next Sunday
this Saturday
3 days
1 month
-2 years
+10 hours
2020-001
2020-W01
2020-W53
2020-W01-7
2020-W53-1
2020-02-29T12:00:00Z
4th of July 2023
1st of January 2023
2nd of February 2024
31st of October 2023
EOF
