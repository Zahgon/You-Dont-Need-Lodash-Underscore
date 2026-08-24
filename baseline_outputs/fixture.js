import { isNaN as isNanAlias } from 'lodash';
import { every } from 'lodash-es';
import lodashIsNaN from 'lodash/isNaN';
import esIsNaN from 'lodash-es/isNaN';
import dotIsNaN from 'lodash.isnan';
import fpTrim from 'lodash/fp/trim';

const { isUndefined: isUndef, keys: keysAlias } = require('lodash');
const { flatten: flattenAlias } = require('lodash/fp');
const { head: headAlias } = require('lodash-es');

let reassigned;
({ contains: reassigned } = require('lodash'));

require('lodash/concat');
require('lodash.lastindexof');
require('lodash-es/startsWith');
require('lodash/fp/endsWith');

const array = [1, 2, [3, 4]];

_.concat(array, 2, [3], [[4]]);
_.first(array);
_.last(array);
_.head(array);
_.flatten(array);
_.startsWith('abc', 'a');
_.endsWith('abc', 'c');
_.trim('  abc  ');
_.isUndefined(2);
_(2).isUndefined();
_([1, 2, [3, 4]]).flatten();
_(2, [3], [[4]]);

lodash.keys({ one: 1, two: 2, three: 3 });
lodash.every(array);

underscore.forEach(array);
underscore.isNaN(NaN);
underscore.contains(array, 1);

array.concat(2, [3], [[4]]);
Object.keys({ one: 1, two: 2, three: 3 });
Number.isNaN(NaN);
[0, 1].forEach();
Object.entries({ one: 1, two: 2 }).forEach();
'abc'.startsWith('a');
'abc'.endsWith('c');
[1, 2, [3, 4]].flat();
