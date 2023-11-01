// Code generated by plangen.

// Copyright 2023 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package queries

var GeneratedColumnPlanTests = []QueryPlanTest{
	{
		Query: `select * from generated_stored_1 where b = 2 order by a`,
		ExpectedPlan: "DescribeQuery(format=tree)\n" +
			" └─ Sort(generated_stored_1.a:0!null ASC nullsFirst)\n" +
			"     └─ IndexedTableAccess(generated_stored_1)\n" +
			"         ├─ index: [generated_stored_1.b]\n" +
			"         ├─ static: [{[2, 2]}]\n" +
			"         └─ Table\n" +
			"             ├─ name: generated_stored_1\n" +
			"             └─ columns: [a b]\n" +
			"",
	},
	{
		Query: `select * from generated_stored_2 where b = 2 and c = 3 order by a`,
		ExpectedPlan: "DescribeQuery(format=tree)\n" +
			" └─ Sort(generated_stored_2.a:0!null ASC nullsFirst)\n" +
			"     └─ IndexedTableAccess(generated_stored_2)\n" +
			"         ├─ index: [generated_stored_2.b,generated_stored_2.c]\n" +
			"         ├─ static: [{[2, 2], [3, 3]}]\n" +
			"         └─ Table\n" +
			"             ├─ name: generated_stored_2\n" +
			"             └─ columns: [a b c]\n" +
			"",
	},
	{
		Query: `delete from generated_stored_2 where b = 3 and c = 4`,
		ExpectedPlan: "DescribeQuery(format=tree)\n" +
			" └─ RowUpdateAccumulator\n" +
			"     └─ Delete\n" +
			"         └─ IndexedTableAccess(generated_stored_2)\n" +
			"             ├─ index: [generated_stored_2.b,generated_stored_2.c]\n" +
			"             ├─ static: [{[3, 3], [4, 4]}]\n" +
			"             └─ Table\n" +
			"                 ├─ name: generated_stored_2\n" +
			"                 └─ columns: [a b c]\n" +
			"",
	},
	{
		Query: `update generated_stored_2 set a = 5, c = 10 where b = 2 and c = 3`,
		ExpectedPlan: "DescribeQuery(format=tree)\n" +
			" └─ RowUpdateAccumulator\n" +
			"     └─ Update\n" +
			"         └─ UpdateSource(SET generated_stored_2.a:0!null = 5 (tinyint),SET generated_stored_2.c:2 = 10 (tinyint),SET generated_stored_2.b:1 = parenthesized((generated_stored_2.a:0!null + 1 (tinyint))))\n" +
			"             └─ IndexedTableAccess(generated_stored_2)\n" +
			"                 ├─ index: [generated_stored_2.b,generated_stored_2.c]\n" +
			"                 ├─ static: [{[2, 2], [3, 3]}]\n" +
			"                 └─ Table\n" +
			"                     ├─ name: generated_stored_2\n" +
			"                     └─ columns: [a b c]\n" +
			"",
	},
	{
		Query: `select * from generated_virtual_1 where c = 7`,
		ExpectedPlan: "DescribeQuery(format=tree)\n" +
			" └─ IndexedTableAccess(generated_virtual_1)\n" +
			"     ├─ index: [generated_virtual_1.c]\n" +
			"     ├─ static: [{[7, 7]}]\n" +
			"     └─ VirtualColumnTable\n" +
			"         ├─ name: generated_virtual_1\n" +
			"         ├─ columns: [generated_virtual_1.a:0!null, generated_virtual_1.b:1, parenthesized((generated_virtual_1.a:0!null + generated_virtual_1.b:1))]\n" +
			"         └─ Table\n" +
			"             ├─ name: generated_virtual_1\n" +
			"             └─ columns: [a b c]\n" +
			"",
	},
	{
		Query: `update generated_virtual_1 set b = 5 where c = 3`,
		ExpectedPlan: "DescribeQuery(format=tree)\n" +
			" └─ RowUpdateAccumulator\n" +
			"     └─ Update\n" +
			"         └─ UpdateSource(SET generated_virtual_1.b:1 = 5 (tinyint),SET generated_virtual_1.c:2 = parenthesized((generated_virtual_1.a:0!null + generated_virtual_1.b:1)))\n" +
			"             └─ IndexedTableAccess(generated_virtual_1)\n" +
			"                 ├─ index: [generated_virtual_1.c]\n" +
			"                 ├─ static: [{[3, 3]}]\n" +
			"                 └─ VirtualColumnTable\n" +
			"                     ├─ name: generated_virtual_1\n" +
			"                     ├─ columns: [generated_virtual_1.a:0!null, generated_virtual_1.b:1, parenthesized((generated_virtual_1.a:0!null + generated_virtual_1.b:1))]\n" +
			"                     └─ Table\n" +
			"                         ├─ name: generated_virtual_1\n" +
			"                         └─ columns: [a b c]\n" +
			"",
	},
	{
		Query: `delete from generated_virtual_1 where c = 6`,
		ExpectedPlan: "DescribeQuery(format=tree)\n" +
			" └─ RowUpdateAccumulator\n" +
			"     └─ Delete\n" +
			"         └─ IndexedTableAccess(generated_virtual_1)\n" +
			"             ├─ index: [generated_virtual_1.c]\n" +
			"             ├─ static: [{[6, 6]}]\n" +
			"             └─ VirtualColumnTable\n" +
			"                 ├─ name: generated_virtual_1\n" +
			"                 ├─ columns: [generated_virtual_1.a:0!null, generated_virtual_1.b:1, parenthesized((generated_virtual_1.a:0!null + generated_virtual_1.b:1))]\n" +
			"                 └─ Table\n" +
			"                     ├─ name: generated_virtual_1\n" +
			"                     └─ columns: [a b c]\n" +
			"",
	},
	{
		Query: `select * from generated_virtual_keyless where v = 2`,
		ExpectedPlan: "DescribeQuery(format=tree)\n" +
			" └─ IndexedTableAccess(generated_virtual_keyless)\n" +
			"     ├─ index: [generated_virtual_keyless.v]\n" +
			"     ├─ static: [{[2, 2]}]\n" +
			"     └─ VirtualColumnTable\n" +
			"         ├─ name: generated_virtual_keyless\n" +
			"         ├─ columns: [generated_virtual_keyless.j:0, parenthesized(json_unquote(json_extract(generated_virtual_keyless.j, '$.a')))]\n" +
			"         └─ Table\n" +
			"             ├─ name: generated_virtual_keyless\n" +
			"             └─ columns: [j v]\n" +
			"",
	},
	{
		Query: `update generated_virtual_keyless set j = '{"a": 5}' where v = 2`,
		ExpectedPlan: "DescribeQuery(format=tree)\n" +
			" └─ RowUpdateAccumulator\n" +
			"     └─ Update\n" +
			"         └─ UpdateSource(SET generated_virtual_keyless.j:0 = {\"a\": 5} (longtext),SET generated_virtual_keyless.v:1 = parenthesized(json_unquote(json_extract(generated_virtual_keyless.j, '$.a'))))\n" +
			"             └─ IndexedTableAccess(generated_virtual_keyless)\n" +
			"                 ├─ index: [generated_virtual_keyless.v]\n" +
			"                 ├─ static: [{[2, 2]}]\n" +
			"                 └─ VirtualColumnTable\n" +
			"                     ├─ name: generated_virtual_keyless\n" +
			"                     ├─ columns: [generated_virtual_keyless.j:0, parenthesized(json_unquote(json_extract(generated_virtual_keyless.j, '$.a')))]\n" +
			"                     └─ Table\n" +
			"                         ├─ name: generated_virtual_keyless\n" +
			"                         └─ columns: [j v]\n" +
			"",
	},
	{
		Query: `delete from generated_virtual_keyless where v = 5`,
		ExpectedPlan: "DescribeQuery(format=tree)\n" +
			" └─ RowUpdateAccumulator\n" +
			"     └─ Delete\n" +
			"         └─ IndexedTableAccess(generated_virtual_keyless)\n" +
			"             ├─ index: [generated_virtual_keyless.v]\n" +
			"             ├─ static: [{[5, 5]}]\n" +
			"             └─ VirtualColumnTable\n" +
			"                 ├─ name: generated_virtual_keyless\n" +
			"                 ├─ columns: [generated_virtual_keyless.j:0, parenthesized(json_unquote(json_extract(generated_virtual_keyless.j, '$.a')))]\n" +
			"                 └─ Table\n" +
			"                     ├─ name: generated_virtual_keyless\n" +
			"                     └─ columns: [j v]\n" +
			"",
	},
}
