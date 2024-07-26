# Python code to
# demonstrate readlines()


file1 = open('results.txt', 'r')
lines = file1.readlines()

# Strips the newline character

placements = []

dnf_position = lines[len(lines)-1]


points = [
    40,
    35,
    32,
    30,
    28,
    26,
    24,
    23,
    22,
    21,
    20,
    19,
    18,
    17,
    16,
    15,
    14,
    13,
    12,
    11,
    10,
    9 ,
    8, 	
    7, 	
    6, 	
    5, 	
    4, 	
    3, 	
    2, 	
    1 
]

for line in lines[:-1]:
    placements.append(line.strip())

count = 0
for placement in placements:
    count += 1
    if count >= int(dnf_position) and dnf_position != 0:
        print("{}: {}\tDNF".format(count, placement.strip()))
    else:
        if count <= len(points):
            print("{}: {}\t+{}P".format(count, placement.strip(), points[count-1]))
        else:
            print("{}: {}".format(count, placement.strip()))

