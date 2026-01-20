select * from business_orders a 
left join business_orders_list b on a.company_id = b.company_id and a.orders_id = b.orders_id
where a.company_id = 218908 and (
	(b.goods_id = 51735 and b.options_id = '0')
	or
	(b.goods_id = 51736 and b.options_id = '1')
)
and (
	a.orders_num like '%DB%'
	or
	a.orders_num like '%2025%'
)
and (
)
and b.goods_id between 0 and 100000

group by b.orders_list_id

limit 0,10;


select * from business_orders a 
left join business_orders_list b on a.company_id = b.company_id and a.orders_id = b.orders_id
where a.company_id = {val "$.params.company_id"} and (
    {range $i, $value := (getValue $ "$.params.goods")}
        {if $i > 0}
           and
        {end}
        (b.goods_id = {val printf("$.params.goods.%d.goods_id", $i)} and b.options_id = {val printf("$.params.goods.%d.options_id", $i)}
	{end}
)
and (
	a.orders_num like '%DB%'
	or
	a.orders_num like '%2025%'
)
and (
)
and b.goods_id between 0 and 100000

group by b.orders_list_id

limit 0,10;